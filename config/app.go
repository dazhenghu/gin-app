package config

type AppConfig struct {
    Addr        string  `yaml:"addr"`          // 监听端口
    Pulic       string  `yaml:"public"`        // 静态资源路径
    ViewBaseDir string  `yaml:"view_base_dir"` // view路径
    Db          DbConfg `yaml:"db"`            // 数据库连接
}
