# go with ueditor

一站式 ueditor go后端

> 虽然官方 ueditor 已经停止维护，但个人觉得现在功能比较完善的前端富文本编辑器中，还是 ueditor 比较好用

## 一、安装

```bash
go get -u github.com/hellflame/ueditor
```

## 二、概念

作为一款前端工具，在使用时首先需要在页面中引入正确的 `资源路径`，该路径包含 `ueditor` 所需的js、css等资源。该 `资源路径` 可直接使用后端提供的文件服务，默认在 `/assets/` 路径下提供 `ueditor` 所需前端资源服务，资源为官方最后一次打包的 `1.4.3.3` 版本

除了 ueditor 在初始化时默认向后端请求的 `配置接口` 外，编辑器的基本编辑功能都无需再请求后端接口(图片和附件等可以通过外部链接引入)。极端情况下，配置接口可直接通过文件服务的形式给出，做到无后端驱动就基本可用的编辑器。

ueditor 的后端实现主要提供了三部分接口，`配置接口`、`文件上传接口` 和 `文件列表接口` 。`配置接口` 主要为接下来的各类文件相关操作提供参数控制；`文件上传接口` 为不同类型的文件上传提供支持，虽然类型不同和控制参数不同，但主要目的是保存到服务器并返回可在页面访问的链接；`文件列表接口` 为编辑器提供已上传文件数据的展示功能，编辑器可服用列表中的文件而不用再次上传。上传后的文件需要能被访问到，所以后端也提供了基本的 `资源服务` 。

## 三、使用

具体使用方法可参考示例 [examples/plain/](examples/plain/)

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

## 四、说明

### 架构

后端接口在设计上主要分为三部分，`UEditor` 、`Storage` 以及 `Bridge` 。

`UEditor` 包含配置与接口数据的处理，并通过 `Storage` 接口进行资源数据的存储与查询，也是一个与 `http` 服务打交道的数据结构 (struct)。其所提供的配置主要为前端编辑器与 `Storage` 提供依据，联系着http服务与存储服务。

`Storage` 定义了存储相关的接口，除了可将资源数据存储到本地外，还可根据该接口协议实现 `minio` 等外部资源的存储。可根据需要进行实现 (如配合 mysql, mongo 等数据库)。

`Bridge` 主要目的为将实现的接口添加到当前 http 接口服务中，由于不同后端框架 (http、gin、beego) 有自己的路由定义和响应方式，需要使用对应的桥接方法，比如示例中的 `BindHTTP` 

### 条件编译

#### 1. 资源路径

该后端服务在编译时默认会将 `资源路径` 进行内嵌，若已使用外部资源，可添加编译条件 `external` 降低发布尺寸。

#### 2. 存储实现

#### 3. 桥接实现


[^cdn]: 如果使用 CDN 或其他 `资源路径` ，此处需要注意 `config.js` 中所给 `serverUrl` 的值需与后端接口保持一致







