Rust

安装racer插件

cargo install racer

使用日更版本
rustup update nightly

# 模式
new type模式
单元结构体 没有字段的结构体
按位复制
悬垂指针
析构函数
值语义


# 内置trait
Copy  按位复制  
Clone 

指针类型：
原生指针


# 语义
引用 &
复制Copy  需要支持按位复制
移动Move
drop
值语义

# 包
包： crate
包管理器  cargo

# 词法作用域：

产生情况
 - let 
 - 花括号{}
 - 函数

Some() 可选值
Ok() 不定值
Box::new(20); 堆内存上存储数字20

# 生命周期忽略规则
每个输入位置上省略的生命周期都将成为一个不同的生命周期参数。
· 如果只有一个输入生命周期的位置（不管是否忽略），则该生命周期都将分配给输出生命周期。
· 如果存在多个输入生命周期的位置，但是其中包含着&self或&mut self，则self的生命周期都将分配给输出生命周期。”

envoy.service.ratelimit.v2.RateLimitService/ShouldRateLimit
