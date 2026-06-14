# maccms10 后台静态资源清单

提取自: https://github.com/magicblack/maccms10
提取时间: 2026-06-15 01:14:43

## 概览

| 资源类型 | 文件数 |
|---------|--------|
| layui 框架 | 104 |
| CSS 样式 | 4 |
| JS 脚本 | 18 |
| 图片资源 | 51 |
| HTML 模板 | 151 |
| **总计** | **329** |

## 目录结构

```
admin-spa/
├── layui/          # layui UI 框架 (104 files)
├── css/            # 公共样式 (4 files)
├── js/             # 公共脚本 (18 files)
├── images/         # 图片资源 (51 files)
└── templates/      # 后台模板 (151 files)
```

## 模板模块详情

### actor (2 files)

- actor/index.html
- actor/info.html

### addon (4 files)

- addon/add.html
- addon/config.html
- addon/index.html
- addon/info.html

### admin (2 files)

- admin/index.html
- admin/info.html

### annex (4 files)

- annex/file.html
- annex/index.html
- annex/info.html
- annex/init.html

### art (3 files)

- art/batch.html
- art/index.html
- art/info.html

### card (2 files)

- card/index.html
- card/info.html

### cash (1 files)

- cash/index.html

### cj (8 files)

- cj/col_content.html
- cj/col_url.html
- cj/index.html
- cj/info.html
- cj/program.html
- cj/publish.html
- cj/show.html
- cj/show_url.html

### collect (9 files)

- collect/actor.html
- collect/art.html
- collect/index.html
- collect/info.html
- collect/role.html
- collect/timing.html
- collect/union.html
- collect/vod.html
- collect/website.html

### comment (4 files)

- comment/blacklist.html
- comment/blacklist_ip.html
- comment/index.html
- comment/info.html

### database (4 files)

- database/export.html
- database/import.html
- database/rep.html
- database/sql.html

### domain (1 files)

- domain/index.html

### extend (23 files)

- extend/editor/ckeditor.html
- extend/editor/kindeditor.html
- extend/editor/tinymce.html
- extend/editor/ueditor.html
- extend/editor/umeditor.html
- extend/email/phpmailer.html
- extend/pay/alipay.html
- extend/pay/codepay.html
- extend/pay/epay.html
- extend/pay/jeepay.html
- extend/pay/weixin.html
- extend/pay/zhapay.html
- extend/sms/aliyun.html
- extend/sms/qcloud.html
- extend/upload/alibaba.html
- extend/upload/ftp.html
- extend/upload/qiniu.html
- extend/upload/s3.html
- extend/upload/uomg.html
- extend/upload/upyun.html
- extend/upload/weibo.html
- extend/urlsend/baidufast.html
- extend/urlsend/baidu.html

### gbook (2 files)

- gbook/index.html
- gbook/info.html

### group (2 files)

- group/index.html
- group/info.html

### images (1 files)

- images/opt.html

### index (4 files)

- index/index.html
- index/login.html
- index/quickmenu.html
- index/welcome.html

### link (2 files)

- link/index.html
- link/info.html

### make (1 files)

- make/opt.html

### manga (3 files)

- manga/batch.html
- manga/index.html
- manga/info.html

### order (1 files)

- order/index.html

### plog (1 files)

- plog/index.html

### public (14 files)

- public/editor.html
- public/empty.html
- public/foot.html
- public/head.html
- public/jump.html
- public/msg.html
- public/pages.html
- public/select_copyright.html
- public/select_hits.html
- public/select_level.html
- public/select_lock.html
- public/select_state.html
- public/select_status.html
- public/select_type.html

### role (2 files)

- role/index.html
- role/info.html

### safety (2 files)

- safety/data.html
- safety/file.html

### system (18 files)

- system/configaisearch.html
- system/configaiseo.html
- system/configapi.html
- system/configassistant.html
- system/configcollect.html
- system/configcomment.html
- system/configconnect.html
- system/configemail.html
- system/config.html
- system/configinterface.html
- system/configpay.html
- system/configplay.html
- system/configseo.html
- system/configsms.html
- system/configupload.html
- system/configurl.html
- system/configuser.html
- system/configweixin.html

### template (4 files)

- template/ads.html
- template/index.html
- template/info.html
- template/wizard.html

### timming (2 files)

- timming/index.html
- timming/info.html

### topic (2 files)

- topic/index.html
- topic/info.html

### type (2 files)

- type/index.html
- type/info.html

### ulog (1 files)

- ulog/index.html

### upload (1 files)

- upload/index.html

### urlsend (1 files)

- urlsend/index.html

### user (3 files)

- user/index.html
- user/info.html
- user/reward.html

### visit (1 files)

- visit/index.html

### vod (4 files)

- vod/batch.html
- vod/index.html
- vod/info.html
- vod/iplot.html

### voddowner (2 files)

- voddowner/index.html
- voddowner/info.html

### vodplayer (3 files)

- vodplayer/import.html
- vodplayer/index.html
- vodplayer/info.html

### vodserver (2 files)

- vodserver/index.html
- vodserver/info.html

### website (3 files)

- website/batch.html
- website/index.html
- website/info.html

## 关键文件

- **系统设置**: templates/system/config.html - 最复杂的页面，包含所有设置 Tab
- **主框架**: templates/index/index.html - 侧边栏 + Tab 布局
- **登录页**: templates/index/login.html
- **视频列表**: templates/vod/index.html
- **视频编辑**: templates/vod/info.html
