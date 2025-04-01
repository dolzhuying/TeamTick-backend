# TeamTick 代码生成器使用说明文档

## 概述

TeamTick 项目使用 oapi-codegen 工具根据 OpenAPI 规范自动生成 Go 语言的 API 接口代码。通过配置文件和模板，我们能够为不同模块生成标准化的接口代码，包括 Gin 框架的路由处理、请求/响应类型定义以及严格模式的接口实现。

## .gen.go 文件内容说明

生成的 .gen.go 文件（如 Auth.gen.go, Groups.gen.go 等）主要包含以下组件：

### 1. 接口定义

每个模块都会定义两种服务器接口：

- 普通接口（如 `AuthServerInterface`、`GroupsServerInterface`）：基于 Gin 的处理函数接口
- 严格接口（如 `AuthStrictServerInterface`、`GroupsStrictServerInterface`）：基于上下文和请求/响应对象的接口

例如：

```go
// 普通接口示例（基于 Gin）
type AuthServerInterface interface {
  // 用户登录
  // (POST /auth/login)
  PostAuthLogin(c *gin.Context)
  // 用户注册
  // (POST /auth/register)
  PostAuthRegister(c *gin.Context)
}

// 严格接口示例（基于上下文和请求/响应对象）
type AuthStrictServerInterface interface {
  // 用户登录
  // (POST /auth/login)
  PostAuthLogin(ctx context.Context, request PostAuthLoginRequestObject) (PostAuthLoginResponseObject, error)
  // 用户注册
  // (POST /auth/register)
  PostAuthRegister(ctx context.Context, request PostAuthRegisterRequestObject) (PostAuthRegisterResponseObject, error)
}
```

### 2. 请求/响应对象

为每个 API 端点定义标准化的请求和响应对象：

- 请求对象：包含路径参数、查询参数和请求体
- 响应对象：包含不同 HTTP 状态码的响应定义

例如：

```go
type PostAuthLoginRequestObject struct {
  Body *PostAuthLoginJSONRequestBody
}

type PostAuthLoginResponseObject interface {
  VisitPostAuthLoginResponse(w http.ResponseWriter) error
}

type PostAuthLogin200JSONResponse struct {
  Code string `json:"code"`
  Data struct {
    Token    *string `json:"token,omitempty"`
    UserId   *int64  `json:"userId,omitempty"`
    Username *string `json:"username,omitempty"`
  } `json:"data"`
}
```

### 3. 路由注册函数

提供将接口处理函数注册到 Gin 路由的函数：

```go
// 注册路由处理函数
func RegisterAuthHandlers(router gin.IRouter, si AuthServerInterface) {
  RegisterAuthHandlersWithOptions(router, si, AuthGinServerOptions{})
}

// 带选项的注册函数
func RegisterAuthHandlersWithOptions(router gin.IRouter, si AuthServerInterface, options AuthGinServerOptions) {
  // 注册路由实现...
}
```

### 4. 中间件包装器

包含用于处理中间件的包装器代码：

```go
type AuthServerInterfaceWrapper struct {
  Handler            AuthServerInterface
  HandlerMiddlewares []AuthMiddlewareFunc
  ErrorHandler       func(*gin.Context, error, int)
}

// 中间件处理函数
func (siw *AuthServerInterfaceWrapper) PostAuthLogin(c *gin.Context) {
  // 中间件执行逻辑...
  siw.Handler.PostAuthLogin(c)
}
```

### 5. 严格模式处理器

提供严格模式接口到 Gin 处理函数的转换：

```go
func NewAuthStrictHandler(ssi AuthStrictServerInterface, middlewares []AuthStrictMiddlewareFunc) AuthServerInterface {
  return &AuthstrictHandler{ssi: ssi, middlewares: middlewares}
}

type AuthstrictHandler struct {
  ssi         AuthStrictServerInterface
  middlewares []AuthStrictMiddlewareFunc
}
```

## 如何在应用中使用生成的代码

### 1. 实现服务器接口

根据你的需求，选择实现普通接口或严格接口：

```go
// 普通接口实现示例
type MyAuthServer struct {
  // 你的依赖注入...
}

func (s *MyAuthServer) PostAuthLogin(c *gin.Context) {
  // 登录逻辑实现
  var req PostAuthLoginJSONRequestBody
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"code": "ERROR", "message": err.Error()})
    return
  }
  // 处理登录逻辑...
  c.JSON(200, PostAuthLogin200JSONResponse{
    Code: "SUCCESS",
    Data: struct{...}{ /* 填充数据 */ },
  })
}

// 严格接口实现示例
type MyAuthStrictServer struct {
  // 你的依赖注入...
}

func (s *MyAuthStrictServer) PostAuthLogin(ctx context.Context, request PostAuthLoginRequestObject) (PostAuthLoginResponseObject, error) {
  // 处理登录逻辑...
  return PostAuthLogin200JSONResponse{
    Code: "SUCCESS",
    Data: struct{...}{ /* 填充数据 */ },
  }, nil
}
```

### 2. 注册路由

在你的 Gin 应用中注册实现好的接口：

```go
func SetupRouter() *gin.Engine {
  router := gin.Default()

  // 使用普通接口
  authServer := &MyAuthServer{}
  RegisterAuthHandlers(router, authServer)

  // 或者使用严格接口
  authStrictServer := &MyAuthStrictServer{}
  strictHandler := NewAuthStrictHandler(authStrictServer, nil)
  RegisterAuthHandlers(router, strictHandler)

  // 注册其他模块...
  groupsServer := &MyGroupsServer{}
  RegisterGroupsHandlers(router, groupsServer)

  return router
}
```

### 3. 添加中间件

你可以为不同的功能模块添加特定的中间件，有两种主要方式：

#### 3.1 为整个模块添加中间件

```go
// 定义认证中间件
authMiddleware := func(c *gin.Context) {
  token := c.GetHeader("Authorization")
  if token == "" {
    c.AbortWithStatusJSON(401, gin.H{"code": "ERROR", "message": "认证失败"})
    return
  }
  // 验证 token...
  c.Next()
}

// 为整个Auth模块添加中间件
authMiddlewares := []AuthMiddlewareFunc{authMiddleware}

// 使用中间件注册路由
RegisterAuthHandlersWithOptions(router, authServer, AuthGinServerOptions{
  Middlewares: authMiddlewares,
})
```

#### 3.2 为不同的路由组添加中间件

你可以创建不同的路由组，并为每组应用不同的中间件：

```go
func SetupRouterWithGroups() *gin.Engine {
  router := gin.Default()

  // 全局中间件
  router.Use(gin.Recovery(), gin.Logger())

  // API 版本分组
  v1 := router.Group("/v1")

  // 认证中间件
  authRequired := func(c *gin.Context) {
    // 验证 token...
  }

  // 管理员权限中间件
  adminRequired := func(c *gin.Context) {
    // 验证管理员权限...
  }

  // 公开 API 路由组（无需认证）
  publicGroup := v1.Group("")
  authServer := &MyAuthServer{}
  RegisterAuthHandlersWithOptions(publicGroup, authServer, AuthGinServerOptions{})

  // 需要认证的 API 路由组
  authGroup := v1.Group("")
  authGroup.Use(authRequired)
  userServer := &MyUserServer{}
  RegisterUsersHandlersWithOptions(authGroup, userServer, UsersGinServerOptions{})

  // 需要管理员权限的 API 路由组
  adminGroup := v1.Group("")
  adminGroup.Use(authRequired, adminRequired)
  groupsServer := &MyGroupsServer{}
  RegisterGroupsHandlersWithOptions(adminGroup, groupsServer, GroupsGinServerOptions{})

  return router
}
```

这种方式允许你根据不同的访问权限或功能模块，为路由分组应用不同的中间件组合。例如：

- 公开 API（登录/注册）无需认证
- 普通用户 API 需要基本认证
- 管理员 API 需要更高权限的认证

你也可以为特定路径添加自定义的前缀：

```go
// API 版本 v2 的路由
v2 := router.Group("/v2")
RegisterAuthHandlersWithOptions(v2, authServer, AuthGinServerOptions{
  BaseURL: "/auth",  // 会注册为 /v2/auth/...
})
```

## 配置文件说明

代码生成通过 config 目录下的 YAML 配置文件控制，每个模块对应一个配置文件：

```yaml
package: gen
generate:
  gin-server: true
  # models: true  # 可选
  strict-server: true
output: ../模块名.gen.go
output-options:
  include-tags: ["模块名"]
  user-templates:
    gin/gin-interface.tmpl: tmpl/gin-interface.tmpl
    gin/gin-wrappers.tmpl: tmpl/gin-wrappers.tmpl
    gin/gin-register.tmpl: tmpl/gin-register.tmpl
    strict/strict-gin.tmpl: tmpl/strict-gin.tmpl
    strict/strict-interface.tmpl: tmpl/strict-interface.tmpl
```

## 优势

1. **标准化**：统一的请求/响应处理模式
2. **类型安全**：完整的类型定义和编译时检查
3. **灵活性**：支持普通和严格两种接口实现方式
4. **中间件支持**：内置中间件机制
5. **文档集成**：代码中包含完整的注释说明，与 OpenAPI 文档保持一致

通过这种代码生成方式，我们可以保持 API 实现的一致性，并减少手动编写重复代码的工作量。
