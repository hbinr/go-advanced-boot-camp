# Go Error处理（二）
这一小节主要探讨如将准确、精简的错误上下文信息透传到上一层，不再重复打印日志和透传重复的错误信息。

为什么要这么做呢？大家可以试想这样一个问题，你调别人提供的接口，其内部是直接`return error`了，然后你可能也会直接 `return error` ，那么别的调用方在调用你提供的接口时，他可能也会直接 `return error` 。

如此反复，那么到了程序的顶层，程序主体将把错误打印到屏幕或日志文件中，错误信息是非常笼统的，并不明确；比如只打印了：`No such file or directory`，并没有明确文件名和目录名，以及具体堆栈信息，这样是不利于排查问题的。

本文主要从以下几个方面来探讨如何解决以上问题：
- 只处理一次错误，不要到处打印日志
- Wrap Error：使用 `github.com/pkg/errors` 库
- Go 1.13 后如何处理：使用 `errors.Is()`、`errors.As`来处理 error 类型的判断
- `github.com/pkg/errors`库和原生 `errors`库如何结合使用



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
        log.Errorf("service: some logic failed,err:%+v",err)
        return err
    }
    // do stuff
}
```

示例代码中，`service`层对 `err` 进行了`两次`错误处理：
- 记录日志：`log.Errorf("some logic failed,err:%+v",err)`
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
    if err != nil{
        log.Errorf("service: some logic failed,err:%+v",err)
        // 业务降级，返回默认的头像
        user.Img = defaultImg 
		return nil // 已经通过业务降级处理了err(记录日志)，那么err就不算错误了，就直接返回nil
    }
    // do stuff
}
```
观察具体例子会发现，我在打印日志的时候都用到了 `%+v`，主要作用是友好的显示错误信息（能够显示字段类型），比较如下：

**直接打印日志：**
```go
fmt.Println("err:", err) // 以下为日志输出
————————————————————————————————————————————————————————————
err: failed to open "test": open test: The system cannot find the file specified.
```
**使用%+v，友好打印日志：**
```go
fmt.Printf("err:%+v", err) // 以下为日志输出
————————————————————————————————————————————————————————————
err:open test: The system cannot find the file specified.
failed to open "test"
go.boot.camp/02-Error-Handling/code.ReadFile
	e:/Workspace/go/src/my-project/go-advanced-boot-camp/02-Error-Handling/code/01_wrap_err_test.go:23
go.boot.camp/02-Error-Handling/code.TestWrapError
	e:/Workspace/go/src/my-project/go-advanced-boot-camp/02-Error-Handling/code/01_wrap_err_test.go:14
testing.tRunner
	E:/Language/go/src/testing/testing.go:1194
runtime.goexit
	E:/Language/go/src/runtime/asm_amd64.s:1371--- PASS: TestWrapError (0.00s)
```

在使用 `%+v`打印日志时，还有个细节值得注意的就是：
- **在程序的顶部或者是工作的 goroutine 顶部(请求入口)，使用 %+v 把堆栈详情记录。**
## Wrap error
接下来讨论很重要的话题：**错误的堆栈信息如何输出**
### 事故复盘
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

在Go中，使用 `fmt.Errorf()` 创建error时，会丢弃原始错误中除文本以外的所有内容，并不会附带堆栈信息，，因为底层调用了error接口的方法 `func Error() string`。

而我们通常开发时都会在日志中增加一定上下文，方便我们定位问题。

因此推荐使用[github.com/pkg/errors](https://github.com/pkg/errors)库来处理error，主要用到了以下方法：
- `New(message string) error`：创建一个error，底层会附带堆栈信息
- `Errorf(format string, args ...interface{})`：创建一个error，可以格式化字符串，底层会附带堆栈信息
- `Wrap(err error, message string) `：包装err，可以自定义上下文信息，底层会附带堆栈信息
- `Wrapf(err error, format string, args ...interface{})`：作用和 `Wrap`类似，可以格式化字符串
- `WithMessage(err error, message string)`：包装err，可以自定义上下文信息，底层不会附带堆栈信息

具体源码如下：
```go
// New returns an error with the supplied message.
// New also records the stack trace at the point it was called. 
func New(message string) error {
	return &fundamental{
		msg:   message,
		stack: callers(),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return &fundamental{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}


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

**示例：**
```go
func TestWrapError(t *testing.T) {
	_, err := ReadFile("test")
	fmt.Printf("err:%+v", err)
}

func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		// Wrapf err： 包含堆栈信息，可以格式化错误内容
		return nil, errors.Wrapf(err, "failed to open %q", path) // %q 单引号围绕的字符字面值，由Go语法安全地转义，这样中文文件名也能正确显示
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		// Wrap err： 包含堆栈信息
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}

```
具体打印的error信息便不展示了。
### 重复 `Wrap()` 的坑
使用 `Wrap()`虽然打印日志时很方便的附带堆栈信息，但使用时也有一个不小的坑：
- 就是**多处 `Wrap()`**，导致打印的错误时会有**多倍的堆栈信息**，因为每 `Wrap()`一次，底层便会调用一次 `withStack()`，就会多输出一次堆栈信息。

如何合理的使用 `Wrap()`呢，给出以下几点建议：
- 1. 在你的**应用代码(指偏向业务的逻辑，不是封装的基础库)** 中，使用 `errors.New` 或者  `errros.Errorf` 返回自定义错误，注意都是指 `pkg/errors`库，如：
```go
func parseArgs(args []string) error {
	if len(args) < 3 {
		return erros.Errorf("not enough arguments, expected at least 3 argument")
	}
	// ...
}
```	
- 2. 如果**调用其他包内的函数（即项目某个函数）**，通常简单的**直接返回err**，如：
```go

if err := somePkg.Logic();err != nil{
	return err
}

```
- 3. 如果是**最底层的业务逻辑，通常是与数据库相关的**，考虑使用 `errors.Wrap` 或者 `errors.Wrapf` **包装**数据库返回的err
- 4. 如果**和第三方库(如github这类库)、标准库、公司或个人封装的基础库进行协作**，考虑使用 `errors.Wrap` 或者 `errors.Wrapf` **包装**这些库返回的err，如：
```go
f, err := os.Open(filePath)
if err != nil {
	return errros.Wrapf(err, "failed to open %q", filePath)
}
```

记不住上面具体建议没关系，记住一个基本原则：**最底层的逻辑需要wrap**，如业务开发是数据库操作相关（Mysql、MongoDB、Redis等），调用基础库，如GO标准库或第三方库