package webhook

import "fmt"

func ProcessWebhookData(data []byte) {
	fmt.Println("🔹 Webhook recebido:")
	fmt.Println(string(data))
}
