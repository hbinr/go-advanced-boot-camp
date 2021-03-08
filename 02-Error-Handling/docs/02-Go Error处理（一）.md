
# Go Error处理（一）

Go 的 error 处理在社区讨论很广泛，使用非常容易，但是想用好却需要考虑很多，大都在吐槽以下的写法：

```go
err := logicOne()
if err != nil{
    // handle err
}
// do something

err = logicTwo()
if err != nil{
    // handle err
}
// do something 


err = logicThree()
if err != nil{
    // handle err
}
// do something 
```

可以看到， `if err != nil` 这种固定的写法占据了较大篇幅，写的多了会很烦，但是又不得不去处理err，有没有一些优雅的方式来处理err呢，这篇文章就重点来探讨**在Go中如何高效优雅的处理Error**。

> error规范约定：error 作为返回值时，最好放置在函数返回值的最好

## 一.Sentinel Error——error与类型错误的变量进行比较
### 1. == 比较
直接进行比较也是一种方式，但是有种硬编码的感觉，必须**事先确定好错误类型或已经知道要发生的错误是什么类型**，这样在错误比较的时候才能处理得当。

让我们通过一个例子来理解这个问题。

`filepath`包的`Glob`函数用于返回与模式匹配的所有文件的名称。当模式出现错误时，该函数将返回一个错误`ErrBadPattern`。

在`filepath`包中定义了`ErrBadPattern`，如下所述：
```go
var ErrBadPattern = errors.New("syntax error in pattern")
```
`errors.New()`用于创建新的错误。模式出现错误时，由Glob函数返回`ErrBadPattern`。

实战看一下就明白:

```go
files, error := filepath.Glob("[")
if error != nil && error == filepath.ErrBadPattern {
	fmt.Println("error:", error)
	return
}
fmt.Println("matched files:", files)
```
输出：
```go
error: syntax error in pattern
```

我们想返回一个匹配 “[” 模式的文件，如果发生错误会与我们预料中的错误类型进行 `==` 比较。如果比较为`True`，则对错误进行处理。

通过输出，我们看到 `error` 确实是`syntax error in pattern`。

但是这种方式有诸多问题：
1. 这些错误，往往是提前约定好的，而且处理起来不太灵活。
2. 最大的问题是引入了外部包，导致在定义 error 和使用 error 的包之间建立了依赖关系，这无疑增加了API的表面积。比如实例中，就引入了`path/filepath`包。当然这是标准库的包，还能接受。如果很多用户自定义的包都定义了错误，那我们就要引入很多包，来判断各种错误，这容易引起循环引用的问题。
3. 上下文

一些有意义的 fmt.Errorf 携带一些上下文，也会破坏调用者的 == ，调用者将被迫查看 error.Error() 方法的输出，以查看它是否与特定的字符串匹配。

不过这种比较的优点就是错误界限比较清楚，能够清晰的知道到底是什么错误
### 2.contains 比较
`contains` 这种方式的比较，是用字符串匹配的方式判断错误字符串里是不是出现了某种错误。
例子如下：
```go
func openFile(path string) error {
	_, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file, err:", err)
	}
	return nil
}

func main(){
	err := openFile("./test.txt")
	if strings.Contains(error.Error(), "not found") {  // error handling
		// handle error
	}
}
```

这种处理方式，强依赖了`error.Error()` 的输出，必须先获取 `error.Error()` 。事实上，我们不应该依赖检测 `error.Error()` 的输出。

Error 方法存在于 error 接口主要用于方便程序员使用，但不是程序(编写测试可能会依赖这个返回)。这个输出的字符串用于记录日志、输出到 stdout 等


## 二.Error Type——断言底层结构类型,并从结构体字段获取更多信息

通过类型断言来判断 error 是哪种类型的错误，通常指的是那些实现了 error 接口的类型。

这些类型一般都是**自定义结构体**，除了error字段外，还有其他字段，提供了额外的信息。

我们看一个实例：
```go
type PathError struct {
	Op   string
	Path string
	Err  error
}

func (e *PathError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }
```
上述代码是 `PathError`类型错误的定义及实现。
`Error()`方法拼接 **操作**、**路径** 和 **实际错误** 并返回它。这样我们就得到了错误信息。

我们验证一下这些字段的输出内容是什么：
```go
f, err := os.Open("./test.txt")
if err, ok := err.(*os.PathError); ok {
	fmt.Printf("err.Op -> %s \n", err.Op)
	fmt.Printf("err.Path -> %s\n", err.Path)
	fmt.Printf("err.Err -> %v\n", err.Err)
	return
}
fmt.Println(f.Name(), "打开成功")
```
输出：
```go
err.Op -> open
err.Path -> ./test.txt
err.Err -> The system cannot find the file specified.
```

通常，使用这样的 error 类型，外层调用者需要使用类型断言来判断错误。

不过错误发生并不一定是自己所希望的那样，具有意外性，如果考虑比较全面，想断言多种类型的错误然后一一处理，会使用很多`if else`或 `switch case`语句。

这样的做的话，无形中会导入很多外部的包，调用者要使用类型断言和类型 switch，就要让自定义的 error 变为 public。这种模型会导致和调用者产生强耦合，从而导致 API 变得脆弱，不太推荐。

尽量避免使用 `error types`，虽然错误类型比 `sentinel errors` 更好，因为它们可以捕获关于出错的更多上下文，但是 error types 共享 error values 许多相同的问题。

建议是避免自定义错误类型，或者至少避免将它们作为公共 API 的一部分。




## 三.断言底层类型的行为

断言底层类型的**行为**，和上一种方式不同，主要判断的是具体行为，体现到代码上基本数就是一些 `method` 。通常指的是调用struct类型的方法来获取更多信息。

举个例子，查看 DNSError 源码：

```go
// DNSError represents a DNS lookup error.
type DNSError struct {
	Err         string // description of the error
	Name        string // name looked for
	Server      string // server used
	IsTimeout   bool   // if true, timed out; not all timeouts set this
	IsTemporary bool   // if true, error is temporary; not all errors set this
	IsNotFound  bool   // if true, host could not be found
}

func (e *DNSError) Timeout() bool { return e.IsTimeout }

func (e *DNSError) Temporary() bool { return e.IsTimeout || e.IsTemporary }
```
从上面的代码中可以看到，`DNSError`有两个方法`Timeout()`和`Temporary()`，它们都返回一个布尔值，表示错误是超时还是临时的。

实战一下：
```go
addrs, err := net.LookupHost("www.bucunzaide.com")

if err != nil {
	if ins, ok := err.(*net.DNSError); ok {
		if ins.IsTimeout {
			fmt.Println("链接超时......")
		} else if ins.IsTemporary {
			fmt.Println("暂时性错误......")
		} else if ins.IsNotFound {
			fmt.Printf("链接无法找到......,err:%v\n", err)
		} else {
			fmt.Println("未知错误......", err)
		}
	}
	return
}
fmt.Println("访问成功，地址为：", addrs)
```
输出：
```go
链接无法找到......,err:lookup www.bucunzaide.com: no such host
```
例子中随便造了一个域名，然后去访问，拿到网络请求返回的 error 后，我们去断言了错误类型，然后去判断是`DNSError`的哪种错误行为，这样我们就能知道请求错误发生的原因了。 

这样做的好处是不需要 import 引用自定义error struct的包，相对前两种方式，推荐使用这种方式，常见于 kit 库、开源的框架等等，比如k8s源码中就大量使用了这种模式。

不过更多的工程师还是偏向于写业务代码，我们通常在写业务的时候，并不可能对每个结构体增加对应的接口及其方法，更常见的是直接将错误透传处理，也就是直接 `return err`，在顶层通过打印 `Error`或 `Info` 级别的日志来记录错误，便于排查问题。

不过这又衍生出另一个问题，如何将准确、精简的错误上下文信息透传到上一层是一件非常值得思考的事情。

一个好的错误信息能够大大提高错误排查的速率，转换过来就是时间成本和资源成本。

在下一节中，便会深入探讨如何将准确、精简的错误上下文信息透传到上一层，不再重复打印日志和透传重复的错误信息。

