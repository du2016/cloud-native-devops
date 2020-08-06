wast是wasm的可读形式

- Emscripten 将llvm转换为js代码
- wabt
- Binaryen


https://webassemblyhub.io/ wasmhub


WASI是一个新的API体系, 由Wasmtime项目设计, 目的是为WASM设计一套引擎无关(engine-indepent), 面向非Web系统(non-Web system-oriented)的API标准. 目前, WASI核心API(WASI Core)在做覆盖文件, 网络等等模块的API, 但这些实现都是刚刚开始实现, 离实用还是有很长路要走.

- wasmer
- wasmtime
- wavm

每个操作系统都会为运行在该系统下的应用程序提供应用程序二进制接口（Application Binary Interface，ABI）

