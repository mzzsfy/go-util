package di

// Package 创建服务包
func Package(providers ...func(Container) error) func(Container) error {
	return func(container Container) error {
		for _, provider := range providers {
			if err := provider(container); err != nil {
				return err
			}
		}
		return nil
	}
}

// LoadPackages 创建包含服务包的容器
func LoadPackages(container Container, packages ...func(Container) error) (Container, error) {
	for _, pkg := range packages {
		if err := pkg(container); err != nil {
			return nil, err
		}
	}
	return container, nil
}
