package model

type WebhookPayload struct {
	APIVersion string `json:"api_version"`
	Event      Event  `json:"event"`
}

type Event struct {
	Aliases                  []string             `json:"aliases"`
	AppID                    string               `json:"app_id"`
	AppUserID                string               `json:"app_user_id"`
	CommissionPercentage     float64              `json:"commission_percentage"`
	CountryCode              string               `json:"country_code"`
	Currency                 string               `json:"currency"`
	EntitlementID            string               `json:"entitlement_id"`
	EntitlementIDs           []string             `json:"entitlement_ids"`
	Environment              string               `json:"environment"`
	EventTimestampMs         int64                `json:"event_timestamp_ms"`
	ExpirationAtMs           int64                `json:"expiration_at_ms"`
	ID                       string               `json:"id"`
	IsFamilyShare            bool                 `json:"is_family_share"`
	OfferCode                string               `json:"offer_code"`
	OriginalAppUserID        string               `json:"original_app_user_id"`
	OriginalTransactionID    string               `json:"original_transaction_id"`
	PeriodType               string               `json:"period_type"`
	PresentedOfferingID      string               `json:"presented_offering_id"`
	Price                    float64              `json:"price"`
	PriceInPurchasedCurrency float64              `json:"price_in_purchased_currency"`
	ProductID                string               `json:"product_id"`
	PurchasedAtMs            int64                `json:"purchased_at_ms"`
	Store                    string               `json:"store"`
	SubscriberAttributes     map[string]Attribute `json:"subscriber_attributes"`
	TakehomePercentage       float64              `json:"takehome_percentage"`
	TaxPercentage            float64              `json:"tax_percentage"`
	TransactionID            string               `json:"transaction_id"`
	Type                     string               `json:"type"`
}

type Attribute struct {
	UpdatedAtMs int64  `json:"updated_at_ms"`
	Value       string `json:"value"`
}
