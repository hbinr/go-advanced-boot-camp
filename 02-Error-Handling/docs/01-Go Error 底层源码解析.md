# Go中error底层结构和创建error源码解析
**Golang**中的错误处理和Java,Python有很大不同，没有`try...catch`语句来处理错误。因此，**Golang**中的错误处理是一个比较有争议的点，如何优雅正确的处理错误是值得去深究的。

今天先记录`error`是什么及如何创建`error`，撸一撸源码。
## 一.初识error
### 1.什么是error
`error`错误指的是可能出现问题的地方出现了问题。比如打开一个文件时失败，这种情况在人们的意料之中 。

而异常指的是不应该出现问题的地方出现了问题。比如引用了空指针，这种情况在人们的意料之外。

可见，错误是业务过程的一部分，而异常不是 。

**Golang**中的错误也是一种类型。错误用内置的`error`类型表示。就像其他类型，如`int`，`float64`等。

错误值可以存储在变量中，也可以从函数中返回，等等。
### 2.error VS exception

Go 的处理异常逻辑是不引入 exception，支持多参数返回，所以你很容易的在函数签名中带上实现了 error interface 的对象，交由调用者来判定。

**如果一个函数返回了 value, error，你不能对这个 value 做任何假设，必须先判定 error**。唯一可以忽略 error 的是，如果你连 value 也不关心。

Go 中有 panic 的机制，如果你认为和其他语言的 exception 一样，那你就错了。当我们抛出异常的时候，相当于你把 exception 扔给了调用者来处理。
比如，你在 C++ 中，把 string 转为 int，如果转换失败，会抛出异常。或者在 java 中转换 string 为 date 失败时，会抛出异常。

**Go panic 意味着 fatal error(就是挂了)。不能假设调用者来解决 panic，意味着代码不能继续运行。**

使用多个返回值和一个简单的约定，Go 解决了让程序员知道什么时候出了问题，并为真正的异常情况保留了 panic。

### 3.error源码

在 `src/builtin/builtin.go` 文件下，定义了错误类型，源码如下：

```go
// src/builtin/builtin.go

// The error built-in interface type is the conventional interface for
// representing an error condition, with the nil value representing no error.
type error interface {
	Error() string
}
```
`error`是一个接口类型，它包含一个 `Error()` 方法，返回值为`string`。任何实现这个接口的类型都可以作为一个错误使用，Error这个方法提供了对错误的描述。

**注意**：error为`nil`代表**没有错误**。


先看一个文件打开错误的例子：
```go
f, err := os.Open("/test.txt")
if err != nil {
	fmt.Println("open failed, err:", err)
	return
}
fmt.Println("file is ：", f)
```
输出：

```go
open failed, err: open /test.txt: The system cannot find the file specified.
```
可以看到输出了具体错误，分别为: 操作`open`，操作对象`/test.txt`，错误原因`The system cannot find the file specified.`

当执行打印错误语句时, fmt 包会自动调用 `err.Error()` 函数来打印字符串。

这就是错误描述是如何在一行中打印出来的原因。

了解了error是什么，我们接下来了解error的创建。
## 二.error创建
创建方式有两种：

 - errors.New()
 - fmt.Errorf()

### 1.errors.New()函数

在`src/errors/errors.go`文件下，定义了 `errors.New()`函数，入参为字符串，返回一个error对象：
```go
// src/errors/errors.go

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string) error {
	return &errorString{text} // 注意：返回的是一个指针类型
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```
`New()`函数返回一个错误，该错误的格式为给定的文本。

其中 `errorString`是一个结构体，只有一个`string`类型的字段s，并且实现了唯一的方法：`Error()`

> 注意：即使文本相同，每次调用 `New()` 也会返回一个不同的error对象，因为是以指针类型为返回值的，指向了不同的地址。


看完了源码，我们具体实战一下：
```go
// 1.errors.New() 创建一个 error
err1 := errors.New("errors.New() 创建的错误")
fmt.Printf("err1 错误类型：%T，错误为：%v\n", err1, err1)
```
输出：

```go
err1 错误类型：*errors.errorString，错误为：errors.New() 创建的错误
```
可以看到，错误类型是 `errorString`指针，前面的`errors.`表明了其在包 `errors` 下。

通常这就够了，它能反映当时出错了，但是有些时候我们需要更加具体的信息。即需要具体的“**上下文**”信息，表明具体的错误值。

这就用到了 `fmt.Errorf` 函数

### 2.fmt.Errorf()函数
`fmt.Errorf()`函数，它先将字符串格式化，并增加上下文的信息，更精确的描述错误。

我们先实战一下，看看和上面的内容有什么不同：
```go
// 2.fmt.Errorf()
err2 := fmt.Errorf("fmt.Errorf() 创建的错误,错误编码为：%d", 404)
fmt.Printf("err2 错误类型：%T，错误为：%v\n", err2, err2)
```
输出：
```go
err2 错误类型：*errors.errorString，错误为：fmt.Errorf() 创建的错误,错误编码为：404
```
可以看到`err2`的类型是`*errors.errorString`，并且错误编码 404 也输出了。。

为什么`err2`返回的错误类型也是 ：`*errors.errorString`，我们不是用 `fmt.Errorf()`创建的吗？

我们先看下`errors.go`中的源码：
```go
// src/fmt/errors.go

func Errorf(format string, a ...interface{}) error {
	p := newPrinter()
	p.wrapErrs = true
	p.doPrintf(format, a)
	s := string(p.buf)
	var err error
	if p.wrappedErr == nil { 
		err = errors.New(s) // 关键点
	} else {
		err = &wrapError{s, p.wrappedErr}
	}
	p.free()
	return err
}
```
通过源码可以看到，`p.wrappedErr` 为 `nil`的时候，会调用`errors.New()`来创建错误。

所以 `err2`的错误类型是`*errors.errorString`这个问题就解答了。

不过又出现了新问题，这个`p.wrappedErr`是什么呢？什么时候为`nil`? 

我们再看个例子：
```go
// 3. go 1.13 新增加的错误处理特性  %w
err3 := fmt.Errorf("err3: %w", err2)  // err3包裹err2错误
fmt.Printf("err3 错误类型：%T，错误为：%v\n", err3, err3)
```
输出：

```go
err3 错误类型：*fmt.wrapError，错误为：err3: fmt.Errorf() 创建的错误,错误编码为：404
```
**注意**：在格式化字符串的时候，有一个 `%w`占位符，表示格式化的内容是一个`error`类型。

我们主要看下`err3`的内容，其包裹了`err2`错误信息，如下：
```go
err3: fmt.Errorf() 创建的错误,错误编码为：404
```

但是`err3`这次是一个 `*fmt.wrapError`类型？这个类型又是源自哪里？怎么会有这样一个类型？又出现了一个新的问题......

好了，带着这些问题，我们从头开始捋一捋源码，就知道它们到底是什么。

在刚刚的源码`fmt.Errorf()`函数中，我们注意到第一行 `p := newPrinter()` 创建了一个 `p` 对象，这个p对象其实就是`pp`结构体指针的实例， `newPrinter()`源码如下：
```go
// src/fmt/print.go

// newPrinter allocates a new pp struct or grabs a cached one.
func newPrinter() *pp {
	p := ppFree.Get().(*pp)
	p.panicking = false
	p.erroring = false
	p.wrapErrs = false
	p.fmt.init(&p.buf)
	return p
}
```
`newPrinter()`函数返回一个 `pp`结构体指针。
 
我们看下这个结构体，并看看`p.wrappedErr`字段在该结构体中定义：
```go
// src/fmt/print.go

// pp is used to store a printer's state and is reused with sync.Pool to avoid allocations.
type pp struct {
    // 省略部分字段
	
	// wrapErrs is set when the format string may contain a %w verb.
	wrapErrs bool
	
	// wrappedErr records the target of the %w verb.
	wrappedErr error
}
```
由于`pp`结构体的字段较多，我们主要看两个字段：

- `wrapErrs `字段，bool类型，当格式字符串包含`％w`动词时，将赋值为true
- `wrappedErr`字段，error类型，记录`％w`动词的目标，即例子的`err2`

所以我们解决了第一问题：`p.wrappedErr`到底是什么，什么时候为`nil`。

> 即：`p.wrappedErr`是 pp 结构体的一个字段，当格式化错误字符串中没有`%w`动词时，其为`nil`。

还有第二个问题， `*fmt.wrapError` 类型源自哪里？

回想一下刚刚的源码，是不是先判断 `if p.wrappedErr == nil { `，然后还有一个`else`语句，也就是说：当`p.wrappedErr`不为`nil`时，执行以下语句：
```go
err = &wrapError{s, p.wrappedErr}
```
`err` 是结构体`wrapError`的实例，其初始化了两个字段，并且是引用取值（前面有`&`）。我们来看看`wrapError`源码：

```go
// src/fmt/errors.go

type wrapError struct {
	msg string
	err error
}

func (e *wrapError) Error() string {
	return e.msg
}

func (e *wrapError) Unwrap() error {
	return e.err
}
```
`wrapError`结构体有两个字段：

 - msg ，`string`类型
 - err，`error`类型

实现了两个方法：
- Error()，也说明`wrapError`结构体实现了 `error`接口，是一个`error`类型
- Unwrap()，作用是返回原错误值，没有自定义的`msg`了。也就是说拆开了一个被包装的 error。

所以我们的第二个问题， `*fmt.wrapError`是什么，就彻底解答了。

至此，捋完`fmt.Errorf()`的源码了，我们了解了想要的内容，至于`p.doPrintf(format, a)`的具体实现内容很复杂，所以就没去深挖了。

## 总结

总结一下吧，**Golang**中创建错误有两种方式：
**第一种**：`errors.New()`函数，其返回值类型为 `*errors.errorString`。

**第二种**：`fmt.Errorf()`函数
当使用`fmt.Errorf()`来创建错误时，核心有以下两点：

1. 错误描述中**不包含** `%w`时，`p.wrappedErr`为`nil`，所以底层也是调用`errors.New()`创建错误。因此错误类型就是`*errors.errorString`。

2. 错误描述中**包含**`%w`时，`p.wrappedErr`不为`nil`，所以底层实例化`wrapError`结构体指针。 因此错误类型是`*fmt.wrapError`，可以理解为包裹错误类型。

但是还有另一个需要研究的点：**error处理**，如何优雅高效的处理error是需要值得思考，这也是Go社区一直在争论的一个话题。

下一节，我们就来学习下在Go中如何处理error。
