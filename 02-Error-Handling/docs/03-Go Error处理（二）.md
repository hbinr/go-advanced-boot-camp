# Go Error处理（二）
这一小节主要探讨如将准确、精简的错误上下文信息透传到上一层，不再重复打印日志和透传重复的错误信息。

为什么要这么做呢？大家可以试想这样一个问题，你调提供的接口，内部直接`return error`了，那么别的调用方在调用你提供的接口时，可能也会直接 `return error` 。

如此反复，那么到了程序的顶层，程序主体将把错误打印到屏幕或日志文件中，错误信息是非常笼统的，并不明确；比如只打印了：`No such file or directory`，并没有明确文件名和目录名，以及具体堆栈信息，这样是不利于排查问题的。

本文主要从以下几个方面来探讨如何解决以上问题：
- Wrap Error：使用 `github.com/pkg/errors` 库
- Go 1.13 后如何处理：使用 `errors.Is()`、`errors.As`来处理 error 类型的判断
- 二者如何结合使用 



## 只处理一次错误，不要到处打印日志
在处理错误时，牢记以以下核心思想：
> You should only handle errors once. Handling an error means inspecting the error value, and making a single decision.

> 你应该只处理一次错误。处理一个错误意味着检查错误值，并做出单一的决定。

看一个 `err` 处理不规范的例子，相信很多人都有类似的写法；
```go
package service 

func SomeLogic() error{
    // ..
    if err := someMethod(); err != nil{
        log.Error("service: some logic failed,err:",err)
        return err
    }
    // do stuff
}
```

示例代码中，`service`层对 `err` 进行了`两次`错误处理：
- 记录日志：`log.Error("some logic failed,err:",err)`
- 返回错误：`return err`

这种方式在Go中很不规范，你应该明确知道如何处理错误，并且是只处理一次，上述示例应该：
- 要么直接 `return err`，不记录日志
- 要么记录日志，然后接其他逻辑，比如业务降级处理

写个伪代码，以账户服务为例，如果获取用户头像失败，那么返回默认头像：
```go
package service 

func GetUserImg(userID int64) error {
    // ..
    user, err := someMethod(userID)
    if err!= nil{
        log.Error("service: some logic failed,err:",err)
        // 业务降级，返回默认的头像
        user.Img = defaultImg 
    }
    // do stuff
}
```

## Wrap error
### 事务复盘
在实际业务开发中，尤其是多层架构中，日志可能到处打印，以传统的三层架构为例：
- service层记录了一次日志
- dao层又记录了一次日志
- controller层也会记录一次日志

这样导致的结果就是日志打印到处都是，影响服务性能，日志落盘的时候是必定耗时且占用内存的。

笔者就遇到过容器的内存因为日志占满导致服务不可用的事故，当时没有人想到是日志的问题，反正是报 OOM了。

于是团队开始查看各指标，发现接口服务正常、worker任务、消息队列、redis服务等也正常，没有积压，然后使用了重启大法，结果服务还是异常，查看日志时，发现并没有任何新日志进来，看了日志配置也正常，最后竟发现日志占比已经100%了，于是清了日志，服务才正常了。

后来复盘时，主要做了以下改善:
- 核心系统增加日志占比报警（短信、邮件），超85%上报
- 各系统负责人每三天检查日志占比，满85%就可以清理
- 各系统负责人查看日志打印情况，去掉不必要的日志，优化现有日志内容
  
第一、二点其实很好做到，但是第三点实施时却并不是那么理想，什么叫不必要的日志，如何优化现有日志内容？优化到什么地步？

毕竟每个系统都有不同的复杂度，工程师的技术能力也不同，优化出来的结果参差不齐，简单说就是没有一个标准，后来也不了了之，因为没有人挨个检查，检查的标准都没有。

虽然之后再没出现上述事故，但是日志的处理还是一个值得思考的问题。
### 使用 `pkg/errors` 库

在使用 `pkg/errors` 库前，先声明以下处理日志的一些约定：
- The error has been logged.  错误要被日志记录。
- The application is back to 100% integrity.  应用程序处理错误，保证100%完整性。
- The current error is not reported any longer.  之后不再报告当前错误。

日志**记录与错误无关且对调试没有帮助的信息应被视为噪音，应予以质疑**。记录的原因是因为某些东西失败了，而日志包含了答案。

在Go中，使用 `log.Info(err)` 这类代码，只会输出error的文本内容（底层调用了error接口的方法 `func Error() string`），并不会附带堆栈信息，而且我们通常开发时都会在日志中增加一定上下文，方便我们定位问题。

因此推荐使用[github.com/pkg/errors](https://github.com/pkg/errors)库来处理error，主要用到了两个方法：
- `Wrap(err error, message string) `：包装err，可以自定义上下文信息，底层会附带堆栈信息
- `WithMessage(err error, message string)`：包装err，可以自定义上下文信息，底层不会附带堆栈信息

具体源码如下：
```go
// errors -> errors.go 

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	err = &withMessage{
		cause: err,
		msg:   message,
	}
	return &withStack{
		err,
		callers(),
	}
}


// WithMessage annotates err with a new message.
// If err is nil, WithMessage returns nil.
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}
```




