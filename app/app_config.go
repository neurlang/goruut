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
