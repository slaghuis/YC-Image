package gateway

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
)

func RegisterWithGateway() {
    regData := map[string]interface{}{
        "type": "images", "name": "images-1",
        "base_url": "http://localhost:8083", "health_url": "http://localhost:8082/healthz",
        "roles": []string{"image"}, "tags": []string{"auth"}, "weight": 1, "prefix": "/api/v1",
    }
    jsonData, _ := json.Marshal(regData)

    req, _ := http.NewRequest("POST", "http://localhost:8080/registry/register", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Registry-Secret", "dev-register-secret")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err == nil {
        defer resp.Body.Close()
        fmt.Println("Registered with Gateway: ", resp.Status)
    }
}
