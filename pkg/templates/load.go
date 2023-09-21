package templates

import "html/template"

type HTMLTemplates struct {
	InitialPurchase *template.Template
	Renewal         *template.Template
	Cancelation     *template.Template
	Expiration      *template.Template
}

const (
	INIT_PURCHASE = "./pkg/templates/init.html"
	RENEWAL       = "./pkg/templates/renew.html"
	CANCELATION   = "./pkg/templates/cancel.html"
	EXPIRATION    = "./pkg/templates/expire.html"
)

var EmailTemplates *HTMLTemplates

func Load() error {
	var err error
	EmailTemplates = &HTMLTemplates{}
	EmailTemplates.InitialPurchase, err = template.ParseFiles(INIT_PURCHASE)
	if err != nil {
		return err
	}
	EmailTemplates.Renewal, err = template.ParseFiles(RENEWAL)
	if err != nil {
		return err
	}
	EmailTemplates.Cancelation, err = template.ParseFiles(CANCELATION)
	if err != nil {
		return err
	}
	EmailTemplates.Expiration, err = template.ParseFiles(EXPIRATION)
	if err != nil {
		return err
	}
	return nil
}
