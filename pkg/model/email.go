package model

type EmailData struct {
	SubscriptionPlan string `json:"subscription_plan"`
	TransactionID    string `json:"transaction_id"`
	PurchaseTime     string `json:"purchase_time"`
	ExpirationTime   string `json:"expiration_time"`
	Price            string `json:"price"`
}
