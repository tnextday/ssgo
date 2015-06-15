ssgo
-----------------

简单的 [SSDB](http://ssdb.io) golang 客户端, 基于 [hissdb](https://github.com/seefan/gossdb)

# 特性

* 包含一个可伸缩的连接池`ConPool`
* 支持 SSDB 的 hash 表与Go结构映射,`Client.MultiH*`函数
* 支持批量命令, `Client.BatchDo`, `ConPool.BatchDo`
* 通用的 SSDB 返回值 `Reply`


# 示例

见```ssdb_test.go```
