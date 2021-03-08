# 一些不规范或不建议的写法

## 一. Indented flow is for errors 缩进的流程是为了防止错误
无错误的正常流程代码，将成为一条直线，而不是缩进的代码。
```go
f, err := os.Open(path)
if err != nil{
    // handle error
}
// do stuff
```

不推荐以下写法，当逻辑较复杂的时候，会缩进更多，代码看起来不简洁优雅
```go

f, err := os.Open(path)
if err == nil{
    // do stuff
}
// handle error
```
## 二. Eliminate error handling by eliminating errors 通过消除错误来消除错误处理

可能有人会有以下的error处理方式：
```go
func SomeLogic(r *Request) error{
    err := someMethod(r.User)
    if err != nil{
        return err
    }
    return nil
}
```

其实上面的代码可以简写为：
```go
func SomeLogic(r *Request) error{
    return someMethod(r.User)
}
```
这样代码开起来简洁优雅很多。