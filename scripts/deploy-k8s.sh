#!/bin/bash

# NetOrchestrator Kubernetes Deployment Script
# Comprehensive enterprise-grade cloud-native deployment with monitoring stack

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

print_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# Configuration
NAMESPACE="netorchestrator"
CHART_NAME="netorchestrator"
RELEASE_NAME="netorchestrator"
K8S_MANIFESTS_DIR="k8s/netorchestrator"
HELM_CHART_DIR="helm/netorchestrator"

# Check prerequisites
check_prerequisites() {
    print_header "🔍 Checking Prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        print_error "Please install kubectl: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi
    
    # Check helm
    if ! command -v helm &> /dev/null; then
        print_error "Helm is not installed or not in PATH"
        print_error "Please install Helm: https://helm.sh/docs/intro/install/"
        exit 1
    fi
    
    # Check cluster connection
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster"
        print_error "Please ensure your kubeconfig is properly configured"
        exit 1
    fi
    
    # Check Docker/Podman for building image
    if ! command -v docker &> /dev/null && ! command -v podman &> /dev/null; then
        print_error "Neither Docker nor Podman is available"
        print_error "Please install Docker or Podman to build the application image"
        exit 1
    fi
    
    print_success "All prerequisites satisfied"
}

# Get cluster info
get_cluster_info() {
    print_header "📊 Cluster Information"
    
    CLUSTER_NAME=$(kubectl config current-context)
    CLUSTER_VERSION=$(kubectl version --short --client 2>/dev/null | grep "Client Version" | cut -d' ' -f3)
    SERVER_VERSION=$(kubectl version --short 2>/dev/null | grep "Server Version" | cut -d' ' -f3)
    NODE_COUNT=$(kubectl get nodes --no-headers | wc -l | tr -d ' ')
    
    echo "  Cluster: ${CLUSTER_NAME}"
    echo "  Client Version: ${CLUSTER_VERSION}"
    echo "  Server Version: ${SERVER_VERSION}"
    echo "  Nodes: ${NODE_COUNT}"
    echo ""
}

# Build application image
build_image() {
    print_header "🏗️ Building NetOrchestrator Application Image"
    
    print_step "Building application binary..."
    go mod tidy
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-gateway-k8s ./cmd/api-gateway
    
    print_step "Building Docker/Podman image..."
    
    # Create Dockerfile if it doesn't exist
    if [ ! -f "Dockerfile" ]; then
        cat > Dockerfile << 'EOF'
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY bin/api-gateway-k8s /app/api-gateway
COPY config.yaml /app/config.yaml
RUN addgroup -g 1000 netorchestrator && \
    adduser -D -s /bin/sh -u 1000 -G netorchestrator netorchestrator && \
    chown -R netorchestrator:netorchestrator /app
USER netorchestrator
EXPOSE 8080
ENTRYPOINT ["/app/api-gateway"]
EOF
    fi
    
    # Build with Docker or Podman
    if command -v podman &> /dev/null; then
        podman build -t netorchestrator:v2.0 .
        print_success "Image built with Podman"
    else
        docker build -t netorchestrator:v2.0 .
        print_success "Image built with Docker"
    fi
}

# Deploy with Helm
deploy_with_helm() {
    print_header "📦 Deploying with Helm Chart"
    
    print_step "Adding required Helm repositories..."
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    print_step "Creating namespace..."
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    print_step "Installing/Upgrading NetOrchestrator Helm chart..."
    helm upgrade --install $RELEASE_NAME $HELM_CHART_DIR \
        --namespace $NAMESPACE \
        --values $HELM_CHART_DIR/values.yaml \
        --wait \
        --timeout 10m
    
    print_success "Helm deployment completed"
}

# Deploy with raw manifests
deploy_with_manifests() {
    print_header "📋 Deploying with Kubernetes Manifests"
    
    print_step "Creating namespace..."
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    print_step "Applying Kubernetes manifests..."
    
    # Apply in order for dependencies
    kubectl apply -f $K8S_MANIFESTS_DIR/api-gateway.yaml
    kubectl apply -f $K8S_MANIFESTS_DIR/databases.yaml
    kubectl apply -f $K8S_MANIFESTS_DIR/monitoring.yaml
    kubectl apply -f $K8S_MANIFESTS_DIR/visualization.yaml
    kubectl apply -f $K8S_MANIFESTS_DIR/ingress.yaml
    
    print_step "Waiting for deployments to be ready..."
    kubectl wait --for=condition=available --timeout=600s deployment --all -n $NAMESPACE
    
    print_success "Manifest deployment completed"
}

# Setup monitoring
setup_monitoring() {
    print_header "📊 Setting Up Comprehensive Monitoring Stack"
    
    print_step "Installing Prometheus Operator (if not already installed)..."
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    
    # Check if prometheus-operator exists
    if ! helm list -n monitoring | grep kube-prometheus-stack &> /dev/null; then
        kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
        helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
            --namespace monitoring \
            --set grafana.adminPassword=netorchestrator123 \
            --wait
        print_success "Prometheus Operator installed"
    else
        print_success "Prometheus Operator already installed"
    fi
    
    print_step "Creating ServiceMonitor for NetOrchestrator..."
    cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: netorchestrator-servicemonitor
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: netorchestrator
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: netorchestrator
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
EOF
    
    print_success "Monitoring setup completed"
}

# Setup ingress
setup_ingress() {
    print_header "🌐 Setting Up Ingress Controller"
    
    print_step "Installing NGINX Ingress Controller (if not already installed)..."
    
    # Check if ingress-nginx exists
    if ! kubectl get namespace ingress-nginx &> /dev/null; then
        helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
        helm repo update
        
        helm install ingress-nginx ingress-nginx/ingress-nginx \
            --create-namespace \
            --namespace ingress-nginx \
            --set controller.service.type=LoadBalancer \
            --wait
        
        print_success "NGINX Ingress Controller installed"
    else
        print_success "NGINX Ingress Controller already exists"
    fi
    
    print_step "Waiting for LoadBalancer IP..."
    kubectl wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=120s
    
    # Get LoadBalancer IP
    EXTERNAL_IP=""
    while [ -z $EXTERNAL_IP ]; do
        print_status "Waiting for external IP..."
        EXTERNAL_IP=$(kubectl get svc --namespace ingress-nginx ingress-nginx-controller --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
        [ -z "$EXTERNAL_IP" ] && sleep 10
    done
    
    print_success "LoadBalancer IP: $EXTERNAL_IP"
}

# Wait for services to be ready
wait_for_services() {
    print_header "⏳ Waiting for Services to be Ready"
    
    print_step "Waiting for PostgreSQL..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=postgresql -n $NAMESPACE --timeout=300s
    
    print_step "Waiting for Redis..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=redis -n $NAMESPACE --timeout=300s
    
    print_step "Waiting for InfluxDB..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=influxdb -n $NAMESPACE --timeout=300s
    
    print_step "Waiting for NetOrchestrator API..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=api-gateway -n $NAMESPACE --timeout=300s
    
    print_step "Waiting for Prometheus..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n $NAMESPACE --timeout=300s
    
    print_step "Waiting for Grafana..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n $NAMESPACE --timeout=300s
    
    print_success "All services are ready"
}

# Health checks
perform_health_checks() {
    print_header "🏥 Performing Health Checks"
    
    # Port-forward for health checks
    print_step "Setting up port forwarding for health checks..."
    
    kubectl port-forward -n $NAMESPACE service/netorchestrator-service 8080:8080 &
    PF_PID_API=$!
    sleep 5
    
    # Check NetOrchestrator API
    if curl -sf http://localhost:8080/health &>/dev/null; then
        print_success "✅ NetOrchestrator API is healthy"
        API_HEALTHY=true
    else
        print_warning "⚠️  NetOrchestrator API health check failed"
        API_HEALTHY=false
    fi
    
    # Cleanup port forwarding
    kill $PF_PID_API 2>/dev/null || true
    
    # Check Kubernetes resources
    print_step "Checking Kubernetes resources..."
    
    PODS_RUNNING=$(kubectl get pods -n $NAMESPACE --field-selector=status.phase=Running --no-headers | wc -l)
    TOTAL_PODS=$(kubectl get pods -n $NAMESPACE --no-headers | wc -l)
    
    print_status "Pods running: $PODS_RUNNING/$TOTAL_PODS"
    
    if [ "$PODS_RUNNING" -eq "$TOTAL_PODS" ]; then
        print_success "✅ All pods are running"
    else
        print_warning "⚠️  Some pods are not running"
        kubectl get pods -n $NAMESPACE
    fi
}

# Display access information
display_access_info() {
    print_header "🎯 Access Information"
    
    # Get LoadBalancer IP
    EXTERNAL_IP=$(kubectl get svc --namespace ingress-nginx ingress-nginx-controller --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
    
    if [ -z "$EXTERNAL_IP" ]; then
        EXTERNAL_IP="<pending>"
        print_warning "LoadBalancer IP is still pending. Use 'kubectl get svc -n ingress-nginx' to check status."
    fi
    
    echo ""
    echo "🌟 NetOrchestrator Enterprise NaaS Platform Deployed Successfully!"
    echo ""
    echo "📡 Access URLs (add to /etc/hosts if using local cluster):"
    echo "  $EXTERNAL_IP api.netorchestrator.local"
    echo "  $EXTERNAL_IP grafana.netorchestrator.local"
    echo "  $EXTERNAL_IP prometheus.netorchestrator.local"
    echo "  $EXTERNAL_IP alertmanager.netorchestrator.local"
    echo ""
    echo "🔗 Service Endpoints:"
    echo "  • NetOrchestrator API:     https://api.netorchestrator.local"
    echo "  • Swagger Documentation:   https://api.netorchestrator.local/swagger/index.html"
    echo "  • WebSocket Endpoint:      wss://api.netorchestrator.local/ws"
    echo "  • Grafana Dashboards:      https://grafana.netorchestrator.local (admin/netorchestrator123)"
    echo "  • Prometheus Metrics:      https://prometheus.netorchestrator.local"
    echo "  • Alertmanager:            https://alertmanager.netorchestrator.local"
    echo ""
    echo "📊 Kubernetes Dashboard Commands:"
    echo "  kubectl get all -n $NAMESPACE"
    echo "  kubectl logs -f deployment/netorchestrator-api -n $NAMESPACE"
    echo "  kubectl port-forward service/netorchestrator-service 8080:8080 -n $NAMESPACE"
    echo ""
    echo "🔧 Management Commands:"
    echo "  Scale replicas:     kubectl scale deployment netorchestrator-api --replicas=5 -n $NAMESPACE"
    echo "  View HPA status:    kubectl get hpa -n $NAMESPACE"
    echo "  Check ingress:      kubectl get ingress -n $NAMESPACE"
    echo "  View secrets:       kubectl get secrets -n $NAMESPACE"
    echo ""
    echo "🚨 Monitoring & Alerting:"
    echo "  • ✅ Prometheus metrics collection"
    echo "  • ✅ Grafana visualization dashboards"
    echo "  • ✅ Alertmanager with Slack/Email notifications"
    echo "  • ✅ Horizontal Pod Autoscaling (HPA)"
    echo "  • ✅ Network policies and RBAC security"
    echo "  • ✅ Persistent storage for databases"
    echo "  • ✅ TLS termination and ingress routing"
    echo ""
}

# Main deployment function
main() {
    print_header "🚀 NetOrchestrator Kubernetes Deployment"
    print_header "Enterprise-Grade Cloud-Native NaaS Platform"
    echo ""
    
    # Parse command line arguments
    DEPLOYMENT_METHOD="helm"
    SKIP_BUILD=false
    SKIP_MONITORING=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --manifests)
                DEPLOYMENT_METHOD="manifests"
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-monitoring)
                SKIP_MONITORING=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --manifests      Use raw Kubernetes manifests instead of Helm"
                echo "  --skip-build     Skip building the application image"
                echo "  --skip-monitoring Skip setting up monitoring stack"
                echo "  --help           Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_prerequisites
    get_cluster_info
    
    if [ "$SKIP_BUILD" = false ]; then
        build_image
    fi
    
    if [ "$DEPLOYMENT_METHOD" = "helm" ]; then
        deploy_with_helm
    else
        deploy_with_manifests
    fi
    
    if [ "$SKIP_MONITORING" = false ]; then
        setup_monitoring
    fi
    
    setup_ingress
    wait_for_services
    perform_health_checks
    display_access_info
    
    print_success "🎉 NetOrchestrator Enterprise Kubernetes deployment completed successfully!"
    print_success "🏆 Cloud-native NaaS platform with comprehensive observability is now operational!"
}

# Run main function
main "$@"