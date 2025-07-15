package app

// GetHttpPort retrieves the HTTP port from the dataset downloads.
func (ac *Configs) GetHttpPort() string {
	for _, config := range ac.Configs {
		port := config.GetHttpPort()

		if port != "" {
			return port
		}
	}
	return ""
}

// GetAdminHttpPort retrieves the HTTP port from the admin configurations.
func (ac *Configs) GetAdminHttpPort() string {
	for _, config := range ac.Configs {
		port := config.GetAdminHttpPort()

		if port != "" {
			return port
		}
	}
	return ""
}

// GetFavIconSite retrieves the favorite icon site from the configurations.
func (ac *Configs) GetFavIconSite() string {
	for _, config := range ac.Configs {
		site := config.GetFavIconSite()

		if site != "" {
			return site
		}
	}
	return ""
}

// GetIpaFlavors retrieves the ipa flavors from the configurations.
func (ac *Configs) GetIpaFlavors() map[string]map[string]string {
	for _, config := range ac.Configs {
		site := config.GetIpaFlavors()

		if site != nil {
			return site
		}
	}
	return nil
}

// GetPolicyMaxWords retrieves the max words per request count policy from the configurations.
func (ac *Configs) GetPolicyMaxWords() int {
	for _, config := range ac.Configs {
		site := config.GetPolicyMaxWords()

		if site != 0 {
			return site
		}
	}
	return 0
}

// GetLoadModels retrieves the models to be loaded from the configurations.
func (ac *Configs) GetLoadModels() []*struct {
	Lang string
	File string
	Size int64
} {
	for _, config := range ac.Configs {
		site := config.GetLoadModels()

		if len(site) > 0 {
			return site
		}
	}
	return nil
}
