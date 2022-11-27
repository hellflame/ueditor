# go with ueditor

一站式 ueditor go后端

API文档: [![GoDoc](https://godoc.org/github.com/hellflame/ueditor?status.svg)](https://godoc.org/github.com/hellflame/ueditor) [![Go Report Card](https://goreportcard.com/badge/github.com/hellflame/ueditor)](https://goreportcard.com/report/github.com/hellflame/ueditor)

> 虽然官方 ueditor 已经停止维护，但个人觉得现在功能比较完善的前端富文本编辑器中，还是 ueditor 比较好用 [^outdate]

受支持环境：

* 路由服务：
  * http
  * [gin](https://github.com/gin-gonic/gin)
  * [gorilla/mux](https://github.com/gorilla/mux)
* 数据存储
  * 本地
  * [minio](https://min.io/)
* 数据索引
  * 本地
  * sqlite
  * [gorm](https://gorm.io/) (orm)

## 一、安装

```bash
go get -u github.com/hellflame/ueditor
```

## 二、概念

作为一款前端工具，在使用时首先需要在页面中引入正确的 `资源路径`，该路径包含 `ueditor` 所需的js、css等资源。该 `资源路径` 可直接使用后端提供的文件服务，默认在 `/assets/` 路径下提供 `ueditor` 所需前端资源服务，资源为官方最后一次打包的 `1.4.3.3` 版本[^version]

除了 ueditor 在初始化时默认向后端请求的 `配置接口` 外，编辑器的基本编辑功能都无需再请求后端接口(图片和附件等可以通过外部链接引入)。极端情况下，配置接口可直接通过文件服务的形式给出，做到无后端驱动就基本可用的编辑器。

ueditor 的后端实现主要提供了三部分接口，`配置接口`、`文件上传接口` 和 `文件列表接口` 。`配置接口` 主要为接下来的各类文件相关操作提供参数控制；`文件上传接口` 为不同类型的文件上传提供支持，虽然上传的类型和限制参数不同，但主要目的还是将文件保存到服务器并返回可在页面访问的链接；`文件列表接口` 为编辑器提供已上传文件数据的展示功能，编辑器可服用列表中的文件而不用再次上传。上传后的文件需要能被访问到，所以后端也提供了基本的 `资源服务` 。

## 三、使用

### 资源引入

如上一节概念中所说，首先需要在页面中使用正确的 `资源路径` ，如默认的 `/assets/` ，并引入编辑器入口文件(ueditor.all.min.js)、配置文件(config.js)和本地化文件(lang/zh-cn/zh-cn.js)，如 [demo.html](examples/plain/demo.html) 中所示:

```html
<script type="text/javascript" src="/assets/config.js"></script>
<script type="text/javascript" src="/assets/ueditor.all.min.js"></script>
<script type="text/javascript" src="/assets/lang/zh-cn/zh-cn.js"></script>
```

可以使用 CDN 替代[^cdn]

### 后端服务

此处使用 go 自带的 `http` 库 + 本地存储 并使用默认配置 作为示例:

```go
// 创建以本地存储作为介质的 editor 实例，文件将存储到本地 uploads 目录中
editor := ueditor.NewEditor(nil, ueditor.NewLocalStorage("uploads"))

// 将后端接口服务与资源服务与默认的 http 服务绑定
ueditor.BindHTTP(nil, nil, editor)
```

[完整示例](examples/plain/serve.go)

具体使用方法可参考示例 [examples/plain/](examples/plain/)

```bash
# 启动示例
go run serve.go
```

启动成功后默认访问 [http://localhost:8080/demo](http://localhost:8080/demo)

#### > 更多后端服务配置

* 绑定路由到 gin

```diff
- ueditor.BindHTTP(nil, nil, editor)
+ router := gin.Default()
+ ueditor.BindGin(router, nil, editor)
```

[完整示例](examples/gin-flavor/serve.go)

* 绑定路由到 gorilla/mux

```diff
- ueditor.BindHTTP(nil, nil, editor)
+ router := mux.NewRouter()
+ ueditor.BindMux(router, nil, editor)
```

[完整示例](examples/mux-flavor/serve.go)

#### > 更多存储服务配置

* 本地存储可以基于sqlite作为数据索引

```go
// 此处选择该sqlite驱动，理论上可以选择支持database/sql的驱动均可
import _ "github.com/mattn/go-sqlite3"

...

// 连接sqlite数据库 resource.db
db, e := sql.Open("sqlite3", "resource.db")
if e != nil {
  panic(e)
}
defer db.Close()

// 创建以本地存储作为介质的 editor 实例，文件将存储到本地 uploads 目录中，使用sqlite作为文件索引
// sqlite 会在这里初始化时自动检查表或创建表
editor := ueditor.NewEditor(nil, ueditor.NewSqliteStorage("uploads", db))

...


```

仅切换绑定到 `UEditor` 的 `Storage` 实现

该方式仅将文件元信息等数据存储到 sqlite 中，文件实体内容依然存储在本地，为避免文件过于集中，此时将资源文件按照 hash 前两个字符分割到不同目录中

* 本地存储可以通过orm框架 gorm 映射mysql、sqlite、psq等数据存储服务存储数据索引

```go
import (
  "github.com/hellflame/ueditor"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
)

...

// 此处使用gorm提供的sqlite驱动
db, _ := gorm.Open(sqlite.Open("resource.db"))

// 通过 NewGormStorage 绑定 gorm 实例
// gorm 会在此时检查表或创建表
editor := ueditor.NewEditor(nil, ueditor.NewGormStorage("uploads", db))

...

```

[完整示例](examples/plain-gorm/serve.go)

* minio 存储

```go
// 使用自己的 minio 实例
client, e := minio.New("192.168.1.8:9000", &minio.Options{
  Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
})
if e != nil {
  panic(e)
}
// create editor with storage
editor := ueditor.NewEditor(nil, ueditor.NewMinioStorage(client))

...
```

[完整示例](examples/plain-minio/serve.go)

此时本地不存储上传的文件，返回给 ueditor 的链接将使用 minio 链接。minio endpoint 将用于生成资源链接，所以需要注意其可访问性

## 四、说明

### 架构

后端接口在设计上主要分为三部分，`UEditor` 、`Storage` 以及 `Bridge` 。

`UEditor` 包含配置与接口数据的处理，并通过 `Storage` 接口进行资源数据的存储与查询，也是一个与 `http` 服务打交道的数据结构 (struct)。其所提供的配置主要为前端编辑器与 `Storage` 提供依据，联系着http服务与存储服务。

`Storage` 定义了存储相关的接口，除了可将资源数据存储到本地外，还可根据该接口协议实现 `minio` 等外部资源的存储。可根据需要进行实现 (如配合 mysql, mongo 等数据库)。

基本的__本地存储__将上传文件存储到不同的类型目录下(图片、文件、视频等)，以文件的md5作为文件名和元数据文件名存储实际的文件内容，同一个文件上传多次只会存储一个副本。原始数据文件与元数据文件需要同时存在才能提供完整的资源服务。

`Bridge` 主要目的为将实现的接口添加到当前 http 接口服务中，由于不同后端框架 (http、gin、mux) 有自己的路由定义和响应方式，需要使用对应的桥接方法，比如示例中的 `BindHTTP` 

### 条件编译

#### 1. 资源路径

该后端服务在编译时默认会将 `资源路径` 进行内嵌，若已使用外部资源[^external]，可添加编译条件 `external` 降低发布尺寸。

```bash
# 取消嵌入资源路径，此时ueditor前端资源需要外部引入
go build -tags external
```

#### 2. 存储实现

存储实现在默认构建时会将所有资源存储代码都打包，如本地存储，sqlite等，若想仅保留一种存储实现：

可首先添加编译条件 `nostorage` 以禁用所有资源存储代码，再添加如下条件选择需要保留的资源存储方式

1. `onlylocal` : 仅保留本地存储 + 本地索引代码
2. `onlysqlite` : 仅保留本地存储 + sqlite 索引代码
2. `onlygorm`：仅保留本地存储 + gorm 框架代码
2. `onlyminio` : 仅保留minio存储

```bash
# 仅保留本地存储 + 本地索引
go build -tags "nostorage onlylocal"
```

#### 3. 桥接实现

后端在构建时默认会将所有桥接代码都进行构建，将 http，gin 等框架均打包进可执行文件中

可首先添加编译条件 `nobridge` 以禁用所有桥接代码，再添加如下条件选择需要保留的框架

1. `onlyhttp` : 仅保留 http 框架代码
2. `onlygin` : 仅保留 gin 框架代码
2. `onlymux` : 仅保留 mux 框架代码

```bash
# 仅保留 gin 框架入口
go build -tags "nobridge onlygin"
```

## 五、示例

为了减少不必要的包依赖，示例服务入口文件默认无法直接构建，可去除头部编译条件 `//go:build ignore` 进行构建，调试时通过 `go run serve.go` 直接运行，如果提示依赖缺失，需要 `go get` 安装依赖

#### [1. http + 纯本地存储](examples/plain)

#### [2. http + 本地存储 + sqlite索引](examples/plain-sqlite)

#### [3. http + 本地存储 + gorm(sqlite)索引](examples/plain-gorm)

#### [4. http + minio](examples/plain-minio)

#### [5. gin + 纯本地存储](examples/gin-flavor)

#### [6. mux + 纯本地存储](examples/mux-flavor)



[^outdate]: 自己在使用时偶尔会修复可见的前端bug
[^cdn]: 如果使用 CDN 或其他 `资源路径` ，此处需要注意 `config.js` 中所给 `serverUrl` 的值需与后端接口保持一致
[^version]: 本来想用最新的开发版 dev-1.5.0，但打包后发现部分功能存在问题，所以还是用了最后一个发布版本

[^external]: 比如使用其他版本的 ueditor 、修复版本或同接口的兼容编辑器





