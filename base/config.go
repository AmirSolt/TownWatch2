package base

type Config struct {
	// TierConfigs map[models.Tier]TierConfig
}

// type TierConfig struct {
// 	Tier     models.Tier
// 	Name     string
// 	Interval string
// 	Amount   int64
// }

func (base *Base) loadConfig() {
	base.Config = &Config{}
}

// 	// m[models.TierT0] = TierConfig{
// 	// 	Tier:     models.TierT0,
// 	// 	Name:     "Free",
// 	// 	Interval: "never",
// 	// 	Amount:   0,
// 	// }
// 	// m[models.TierT1] = TierConfig{
// 	// 	Tier:     models.TierT1,
// 	// 	Name:     "Monthly",
// 	// 	Interval: "month",
// 	// 	Amount:   1000,
// 	// }
// 	// m[models.TierT2] = TierConfig{
// 	// 	Tier:     models.TierT2,
// 	// 	Name:     "Yearly",
// 	// 	Interval: "year",
// 	// 	Amount:   10000,
// 	// }

// 	base.Config = &Config{}
// }
