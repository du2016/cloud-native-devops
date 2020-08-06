# wasm介绍

wasm 是一个可移植、体积小、加载快并且兼容 Web 的全新格式.

wasm 代码格式：
- .wasm - wasm二进制格式

- .wabt - web assembly text format，编译结果的文本格式，用于调试

通过[wabt](https://github.com/WebAssembly/wabt)可以实现wasm和wabt格式的转换

# wasi介绍 

WASI是一个新的API体系, 由Wasmtime项目设计, 目的是为WASM设计一套引擎无关(engine-indepent), 面向非Web系统(non-Web system-oriented)的API标准. 目前, WASI核心API(WASI Core)在做覆盖文件, 网络等等模块的API, 但这些实现都是刚刚开始实现, 离实用还是有很长路要走.

目前支持wasi的运行时有以下几种：

- wasmer
- wasmtime
- wavm

# wavm介绍

WAVM是WebAssembly虚拟机，设计用于非Web应用程序。

# 特点

- 快速

WAVM使用LLVM将WebAssembly代码编译为具有接近本机性能的机器代码。在某些情况下，它甚至可以胜过本机性能，这要归功于它能够生成针对运行代码的确切CPU进行了调整的机器代码。

WAVM还利用虚拟内存和信号处理程序来执行WebAssembly的边界检查的内存访问，其成本与本机的未经检查的内存访问相同。

- 安全

WAVM阻止WebAssembly代码访问WebAssembly虚拟机*之外的状态，或调用未与WebAssembly模块明确链接的本机代码。

# 安装

对于centos可以使用官方预编译rpm包进行安装

```
yum install -y https://github.com/WAVM/WAVM/releases/download/nightly%2F2020-05-28/wavm-0.0.0-prerelease-linux.rpm
```

# 用法示例

clone官方库
```
git clone https://github.com/WAVM/WAVM
cd Examples
```

## 运行官方示例程序

```
wavm run helloworld.wast
wavm run zlib.wasm
wavm run trap.wast
wavm run echo.wast "Hello, world!"
wavm run helloworld.wast | wavm run tee.wast
wavm run --enable simd blake2b.wast
```

## 拆解wasm为wast

disassemble通过disassemble可以将wasm拆解为wast可读格式

```
wavm disassemble zlib.wasm zlib.wast
```


## 设置cache

WAVM_OBJECT_CACHE_DIR 环境变量为wavm设置运行时缓存

```
export WAVM_OBJECT_CACHE_DIR=/path/to/existing/directory
wavm run huge.wasm # Slow
wavm run huge.wasm # Fast
```


# 使用rust 实现wasi规范的wasm程序

## 查看rust支持的目标

通过执行

```
rustup target list
```

- asmjs-unknown-emscripten 通过emscripten 工具链编译为asmjs，asmjs也是为了解决js性能问题
- wasm32-unknown-unknown。此目标直接使用 llvm 后端编译成 wasm。 它适合纯 rust 代码编译，譬如你没有 C 依赖的时候。 跟 emscripten 目标比起来，它默认就生成更加洗练的代码， 而且也便于设置搭建。此处查看如何设置搭建.
- wasm32-unknown-emscripten。此目标利用 emscripten 工具链编译成 wasm。 当你具有 C 依赖的时候就得使用它了，包括 libc
- wasm32-wasi wasi规范的目标

## 创建rust lib项目

创建项目

```
cargo new --lib testwasi
```

## 项目配置

Cargo.toml中lib的配置如下

```
[lib]
name = "testwasi"
path = "src/lib.rs"
crate-type =["cdylib"]
```

## 代码实现

```
// # 直接把把名字作为符号写到目标文件中
#[no_mangle]
pub extern fn test(a: i32, b: i32) {
    let z = a + b;
    println!("The value of x is: {}", z);
}
```

## 编译

.cargo/config添加以下内容，制定编译结果为wasi格式

```
[build]
target = "wasm32-wasi"
```

执行`cargo build`

或执行`cargo build --target=wasm32-wasi`

## 使用wavm运行rust编译的wasm程序

```
# wavm run --function=test  --abi=wasi target/wasm32-wasi/debug/testwasi.wasm 1 2
The value of x is: 3
```

# 总结

wasm虽然一开始是为了解决js的性能问题，但是由于其高性能，跨平台，众多运行时支持，已经不局限于web端，走向服务端，现在已经应用于servicemesh、serverless等方向,个人认为其可能成为下一代的container，相信其未来必定有更广泛的应用场景。


扫描关注我:

![微信](http://img.rocdu.top/20200527/qrcode_for_gh_7457c3b1bfab_258.jpg)