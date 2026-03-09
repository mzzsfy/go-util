// Package di 提供服务包组合功能
// 服务包允许将多个服务提供者组合成一个单元进行批量注册
package di

// Package 创建服务包
// 将多个服务提供者函数组合成一个包函数
// 便于批量注册相关服务
// 参数:
//   - providers: 服务提供者函数列表
// 返回:
//   - 组合后的包函数，执行时会依次注册所有服务
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
// 将多个服务包加载到容器中
// 参数:
//   - container: 目标容器
//   - packages: 服务包函数列表
// 返回:
//   - 加载完成后的容器
//   - 如果任何包加载失败，返回错误
func LoadPackages(container Container, packages ...func(Container) error) (Container, error) {
	for _, pkg := range packages {
		if err := pkg(container); err != nil {
			return nil, err
		}
	}
	return container, nil
}
