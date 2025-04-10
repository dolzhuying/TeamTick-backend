### DAL 层使用

```
|--dal  对底层数据库封装，包括ORM操作实现、数据模型、数据库连接
  |--dao  ORM操作封装
    |--impl  不同模块的ORM操作具体实现，实现 interface.go 中的接口
    |--main.go 封装DAO对象的工厂函数
    |--interface.go  提供不同模块 ORM 操作的接口
    |--transaction.go  提供事务支持
  |--models 数据模型，对数据库模型结构化
  |-- db.go  提供连接数据库的方法

```

三层架构： `controllers--service--dao`

调用示例：

```go
func (s*UserService) CreateUser(ctx context.Context,username,passowrd string){
    return dao.WithTransaction(ctx,func(tx *gorm,DB)error{
        user:=models.User{
            username:username,
            password:password,
        }
        if err:=s.userDao.Create(ctx,&user,tx);err!=nil{
            return err
        }
        return nil

    })
}
```

在上述代码中，dao.WithTransaction 以函数式的形式自动开启事务并管理事务的提交和失败回滚，无需手动管理事务的开启、提交、回滚，较为便捷安全

引入 `context`：

```go
ctx context.Context
```

该 `context` 来源于 `controllers` 接受的 `gin.Context` 提取出的标准化 `context`：`ctx.Request.Context()`，向 `dao` 层传入的主要目的在于对数据库操作进行超时控制和取消，防止客户端断联但是服务端仍在进行数据库操作，或长时间数据库操作占用连接，降低吞吐量（可在 `controllers` 层对 `ctx` 设置超时时间）
