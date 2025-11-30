开发语言：golang

目标：封装一个全局上下文库

包：goocontext

目录：goo-context

实现功能：

1. 集成context.Context
2. 可以设置key, value; 可以获取value值
3. 基于2，设置全局AppName, TraceId
4. 上下文要控制退出，比如WithCancel
5. 上下文要控制信号处理，比如WithSignalNotify
6. 上下文要控制超时，比如WithTimeout
7. 可以提供基于gin.context的上下文处理
8. 可以提供基于grpc context的上下午处理
9. value.go 要支持 string, int, int32, int64, float32, float64, bool 相关类型的方法数据返回，返回时要自动换行类型，比如：
   1. ValueInt64() 返回时，自动转换 string int int32 float32 float64 bool
   2. ValueString() 返回时，自动转换 int int32 int64 float32 float64 bool
   3. boolString() 返回时，0 nil null等返回false，其他情况返回true
