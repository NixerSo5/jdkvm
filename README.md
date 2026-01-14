# JDKVM - Java Version Manager

JDKVM是一个用于在Windows系统上管理多个Java版本的工具，参考了NVM-Windows的设计模式，使用Go语言开发。

## 目录结构
```
jdkvm/
└── src/                    # 源代码目录
    ├── arch/              # 架构检测和验证
    ├── file/              # 文件操作
    ├── java/              # Java版本管理
    ├── utility/           # 实用工具
    ├── web/               # 网络下载功能
    ├── go.mod             # Go模块配置
    └── jdkvm.go            # 主入口文件
```

## 安装和使用步骤

### 1. 安装Go语言环境

在使用JDKVM之前，需要在系统上安装Go语言环境：

1. 访问Go官方网站：https://golang.org/dl/
2. 下载适合Windows的Go安装包（msi文件）
3. 运行安装程序，按照提示完成安装
4. 安装完成后，打开命令行工具，输入`go version`验证安装是否成功

### 2. 构建JDKVM项目

1. 进入JDKVM项目的src目录：
   ```bash
   cd jdkvm\src
   ```

2. 安装项目依赖：
   ```bash
   go mod tidy
   ```

3. 构建JDKVM可执行文件：
   ```bash
   go build -o jdkvm.exe
   ```

4. 构建成功后，当前目录会生成`jdkvm.exe`文件


### 3. 使用JDKVM命令

设置完成后，可以在命令行中使用以下JDKVM命令：

#### 安装Java版本
```bash
jdkvm install 17.0.11  # 安装Java 17.0.11
jdkvm install 11.0.23  # 安装Java 11.0.23
jdkvm install 8.0.412  # 安装Java 8.0.412
```

#### 切换Java版本
```bash
jdkvm use 17.0.11  # 使用Java 17.0.11
jdkvm use 11.0.23  # 使用Java 11.0.23
```

#### 列出已安装的Java版本
```bash
jdkvm list  # 或 jdkvm ls
jdkvm list installed  # 列出已安装的版本
jdkvm list available  # 列出可用的版本（需要网络连接）
```

#### 查看当前使用的Java版本
```bash
jdkvm current
```

#### 卸载Java版本
```bash
jdkvm uninstall 8.0.412  # 或 jdkvm rm 8.0.412
```

#### 查看JDKVM版本
```bash
jdkvm version
```


### 5. 配置Java镜像源（可选）

如果默认的Java下载源速度较慢，可以配置自定义镜像源：

```bash
jdkvm java_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium/
```

## 注意事项

1. **管理员权限**：某些操作（如创建符号链接）可能需要管理员权限，建议以管理员身份运行命令行工具

2. **Java版本格式**：请使用完整的版本号格式，如`17.0.11`、`11.0.23`、`8.0.412`

3. **网络连接**：安装Java版本时需要网络连接来下载安装包

4. **环境变量**：确保正确设置了所有必要的环境变量，否则JDKVM可能无法正常工作

## 常见问题

### Q: 为什么`jdkvm install`命令提示下载失败？
A: 可能的原因包括网络连接问题、代理设置错误或版本号不正确。请检查网络连接和版本号格式，确保版本号存在于Adoptium仓库中。

### Q: 为什么切换版本后`java -version`显示的版本没有变化？
A: 可能是环境变量设置不正确，或者需要重启命令行窗口使环境变量生效。请检查`JDKVM_SYMLINK`和`PATH`环境变量是否正确设置。

### Q: 为什么创建符号链接失败？
A: 创建符号链接需要管理员权限，请以管理员身份运行命令行工具。



## 许可证

MIT License
