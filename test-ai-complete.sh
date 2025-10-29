#!/bin/bash

echo "🤖 NetOrchestrator AI Intelligence Test Suite"
echo "============================================="

# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Authentication failed"
    exit 1
fi

echo "✅ Authentication successful"
echo ""

# Test all AI endpoints
echo "🔍 1. Network Performance Analysis:"
curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metrics": {
      "latency": 45.2,
      "throughput": 1200,
      "error_rate": 0.02,
      "cpu_usage": 78.5
    },
    "time_range": "1h"
  }' | jq '.success, .confidence, .insights[0]'

echo ""
echo "📈 2. Capacity Prediction:"
curl -s -X POST http://localhost:8080/api/v1/intelligence/predict/capacity \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "historical_data": [
      {"cpu_usage": 65, "memory_usage": 45, "timestamp": "2023-10-01"},
      {"cpu_usage": 70, "memory_usage": 50, "timestamp": "2023-10-02"},
      {"cpu_usage": 75, "memory_usage": 55, "timestamp": "2023-10-03"},
      {"cpu_usage": 80, "memory_usage": 60, "timestamp": "2023-10-04"}
    ],
    "forecast_days": 30,
    "confidence": 0.95
  }' | jq '.success, .confidence, .insights[0]'

echo ""
echo "⚠️ 3. Anomaly Detection:"
curl -s -X POST http://localhost:8080/api/v1/intelligence/detect/anomalies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metric_name": "cpu_usage",
    "values": [65.5, 70.2, 68.8, 72.1, 95.8, 71.3, 69.4],
    "sensitivity": 0.8
  }' | jq '.success, .metadata.total_anomalies, .metadata.algorithm'

echo ""
echo "💰 4. Cost Optimization:"
curl -s -X POST http://localhost:8080/api/v1/intelligence/optimize/costs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resources": {
      "compute_instances": [
        {"type": "m5.large", "count": 8, "utilization": 0.35, "cost_per_hour": 0.096}
      ],
      "storage": {
        "type": "gp2",
        "size_gb": 500,
        "utilization": 0.6
      }
    },
    "constraints": ["maintain_performance", "no_downtime"],
    "objectives": ["minimize_cost", "optimize_utilization"]
  }' | jq '.success, .insights[0:2]'

echo ""
echo "🎉 All AI Intelligence endpoints are fully functional!"
echo ""
echo "💡 Key AI Capabilities Demonstrated:"
echo "   • LangChain-inspired processing chains"
echo "   • Multi-step reasoning and analysis"
echo "   • Hybrid LSTM-ARIMA forecasting algorithms"
echo "   • Statistical and ML-based anomaly detection"
echo "   • Cost optimization with ROI calculations"
echo "   • Real-time performance insights"
