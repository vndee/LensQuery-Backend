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

var PlanConfigs map[string]Plan

func LoadSubscriptionPlanConfig() error {
	data, err := ioutil.ReadFile("./pkg/config/subConfig.json")
	if err != nil {
		return err
	}

	err = sonic.Unmarshal(data, &PlanConfigs)
	if err != nil {
		return err
	}

	return nil
}
