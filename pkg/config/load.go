package config

import (
	"io/ioutil"

	"github.com/bytedance/sonic"
)

type Plan struct {
	CustomLLMProvider  bool   `json:"CustomLLMProvider"`
	EquationOCRSnap    int    `json:"EquationOCRSnap"`
	FullChatExperience bool   `json:"FullChatExperience"`
	TextOCRSnap        int    `json:"TextOCRSnap"`
	Name               string `json:"name"`
}

type Package struct {
	RCBronze int `json:"rc_bronze"`
	RCSilver int `json:"rc_silver"`
	RCGold   int `json:"rc_gold"`
}

var StorePackages *Package
var AppStorePlanConfigs map[string]Plan
var PlayStorePlanConfigs map[string]Plan

func loadAppStoreSubscriptionPlanConfig() error {
	data, err := ioutil.ReadFile("./pkg/config/appstore.json")
	if err != nil {
		return err
	}

	err = sonic.Unmarshal(data, &AppStorePlanConfigs)
	if err != nil {
		return err
	}

	return nil
}

func loadPlayStoreSubscriptionPlanConfig() error {
	data, err := ioutil.ReadFile("./pkg/config/playstore.json")
	if err != nil {
		return err
	}

	err = sonic.Unmarshal(data, &PlayStorePlanConfigs)
	if err != nil {
		return err
	}

	return nil
}

func LoadSubscriptionPlanConfig() error {
	err := loadAppStoreSubscriptionPlanConfig()
	if err != nil {
		return err
	}

	err = loadPlayStoreSubscriptionPlanConfig()
	if err != nil {
		return err
	}

	return nil
}

func LoadStorePackagesConfig() error {
	data, err := ioutil.ReadFile("./pkg/config/packages.json")
	if err != nil {
		return err
	}

	err = sonic.Unmarshal(data, &StorePackages)
	if err != nil {
		return err
	}

	return nil
}
