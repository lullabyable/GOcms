/**
 * GOcms SPA 核心函数库
 * 统一 API 调用、表格初始化、表单提交等辅助功能
 */

// 统一 API 调用
var API = {
    _handle401: function(res) {
        if (res.status === 401) {
            window.top.location.href = '/admin/login';
            throw new Error('未登录');
        }
        return res;
    },
    get: function(url) {
        return fetch(url, {
            headers: { 'Accept': 'application/json' },
            credentials: 'same-origin'
        }).then(function(res) { return API._handle401(res); })
          .then(function(res) { return res.json(); });
    },
    post: function(url, data) {
        return fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            credentials: 'same-origin',
            body: JSON.stringify(data || {})
        }).then(function(res) { return API._handle401(res); })
          .then(function(res) { return res.json(); });
    },
    form: function(url, formData) {
        return fetch(url, {
            method: 'POST',
            body: formData,
            credentials: 'same-origin',
            headers: { 'Accept': 'application/json' }
        }).then(function(res) { return API._handle401(res); })
          .then(function(res) { return res.json(); });
    }
};

// URL 参数解析
function getQueryParam(name) {
    var reg = new RegExp('(^|&)' + name + '=([^&]*)(&|$)', 'i');
    var r = window.location.search.substr(1).match(reg);
    if (r != null) return decodeURIComponent(r[2]);
    return null;
}

// 获取所有 URL 参数
function getAllParams() {
    var params = {};
    var search = window.location.search.substring(1);
    if (search) {
        search.split('&').forEach(function(pair) {
            var kv = pair.split('=');
            if (kv.length === 2) {
                params[kv[0]] = decodeURIComponent(kv[1]);
            }
        });
    }
    return params;
}

// layui 表格初始化辅助
function initTable(config) {
    layui.use(['table', 'laypage', 'layer', 'form'], function() {
        var table = layui.table,
            laypage = layui.laypage,
            layer = layui.layer,
            form = layui.form,
            $ = layui.jquery;

        // 全选/反选
        form.on('checkbox(allChoose)', function(data) {
            var child = $(data.elem).parents('table').find('tbody input[type="checkbox"]');
            child.each(function(index, item) {
                item.checked = data.elem.checked;
            });
            form.render('checkbox');
        });

        // 状态切换
        form.on('switch(switchStatus)', function(data) {
            var url = $(this).attr('data-href');
            if (url) {
                API.post(url, { value: data.elem.checked ? 1 : 0 }).then(function(r) {
                    layer.msg(r.msg || '操作成功', {time: 1500});
                });
            }
        });

        // 分页
        if (config.total > 0) {
            laypage.render({
                elem: 'pages',
                count: config.total,
                limit: config.limit || 20,
                curr: config.page || 1,
                layout: ['count', 'prev', 'page', 'next', 'limit', 'skip'],
                jump: function(obj, first) {
                    if (!first && config.onPageChange) {
                        config.onPageChange(obj.curr, obj.limit);
                    }
                }
            });
        }

        // 搜索
        $(document).on('click', '.j-search', function() {
            var form = $(this).closest('form');
            var params = {};
            form.serializeArray().forEach(function(item) {
                if (item.value !== '') {
                    params[item.name] = item.value;
                }
            });
            if (config.onSearch) {
                config.onSearch(params);
            }
        });

        // 删除
        $(document).on('click', '.j-tr-del', function() {
            var url = $(this).attr('data-href');
            if (url) {
                layer.confirm('确定要删除吗？', function(index) {
                    API.post(url).then(function(r) {
                        layer.msg(r.msg || '操作成功', {time: 1500});
                        if (r.code === 1) {
                            setTimeout(function() { location.reload(); }, 1500);
                        }
                    });
                    layer.close(index);
                });
            }
        });

        // 批量操作
        $(document).on('click', '.j-page-btns', function() {
            var url = $(this).attr('data-href');
            var ids = [];
            $('input.checkbox-ids:checked').each(function() {
                ids.push($(this).val());
            });
            if (ids.length === 0 && $(this).attr('data-checkbox') !== 'no') {
                layer.msg('请先选择数据');
                return;
            }
            layer.confirm('确定要执行此操作吗？', function(index) {
                API.post(url, { ids: ids.join(',') }).then(function(r) {
                    layer.msg(r.msg || '操作成功', {time: 1500});
                    if (r.code === 1) {
                        setTimeout(function() { location.reload(); }, 1500);
                    }
                });
                layer.close(index);
            });
        });
    });
}

// 表单提交辅助
function bindForm(formId, submitUrl, successCallback) {
    layui.use(['form', 'layer'], function() {
        var form = layui.form,
            layer = layui.layer,
            $ = layui.jquery;

        form.on('submit(formSubmit)', function(data) {
            layer.msg('正在提交...', {time: 500000});
            var formData = data.field;

            // 处理数组字段
            var formEl = document.getElementById(formId || 'editForm');
            if (formEl) {
                var arrayFields = {};
                $(formEl).serializeArray().forEach(function(item) {
                    if (item.name.endsWith('[]')) {
                        var key = item.name.slice(0, -2);
                        if (!arrayFields[key]) arrayFields[key] = [];
                        arrayFields[key].push(item.value);
                    }
                });
                Object.keys(arrayFields).forEach(function(key) {
                    formData[key] = arrayFields[key];
                });
            }

            API.post(submitUrl, formData).then(function(r) {
                layer.closeAll();
                if (r.code === 1) {
                    layer.msg(r.msg || '保存成功', {time: 1500});
                    if (successCallback) {
                        successCallback(r);
                    }
                } else {
                    layer.msg(r.msg || '保存失败', {time: 2000});
                }
            }).catch(function(err) {
                layer.closeAll();
                layer.msg('请求失败: ' + err.message, {time: 2000});
            });

            return false;
        });
    });
}

// 加载数据到表单
function loadFormData(url, formId) {
    return API.get(url).then(function(r) {
        if (r.code === 1 && r.data) {
            var data = r.data;
            var form = document.getElementById(formId || 'editForm');
            if (!form) return r;

            Object.keys(data).forEach(function(key) {
                var el = form.querySelector('[name="' + key + '"]');
                if (el) {
                    if (el.type === 'checkbox') {
                        el.checked = !!data[key];
                    } else if (el.type === 'radio') {
                        var radios = form.querySelectorAll('[name="' + key + '"]');
                        radios.forEach(function(radio) {
                            radio.checked = (radio.value == data[key]);
                        });
                    } else {
                        el.value = data[key] || '';
                    }
                }
            });

            // 重新渲染 layui 表单
            layui.use('form', function() {
                layui.form.render();
            });
        }
        return r;
    });
}

// 填充 select 下拉框
function fillSelect(selectEl, options, valueField, labelField, selectedValue) {
    var html = '<option value="">请选择</option>';
    if (Array.isArray(options)) {
        options.forEach(function(opt) {
            var val = typeof opt === 'object' ? opt[valueField] : opt;
            var label = typeof opt === 'object' ? opt[labelField] : opt;
            var selected = (val == selectedValue) ? ' selected' : '';
            html += '<option value="' + val + '"' + selected + '>' + label + '</option>';
        });
    }
    if (typeof selectEl === 'string') {
        document.querySelector(selectEl).innerHTML = html;
    } else {
        selectEl.innerHTML = html;
    }
    layui.use('form', function() {
        layui.form.render('select');
    });
}

// 生成随机数
function rndNum(min, max) {
    if (!max) { max = min; min = 0; }
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

// 图片 URL 处理
function mac_url_img(url) {
    if (!url) return '';
    if (url.indexOf('://') > -1) return url;
    return ROOT_PATH + '/' + url;
}

// 时间格式化
function mac_day(timestamp, showColor) {
    if (!timestamp) return '';
    var date = new Date(timestamp * 1000);
    var now = new Date();
    var diff = now - date;
    var day = 24 * 60 * 60 * 1000;

    var y = date.getFullYear();
    var m = ('0' + (date.getMonth() + 1)).slice(-2);
    var d = ('0' + date.getDate()).slice(-2);
    var h = ('0' + date.getHours()).slice(-2);
    var mi = ('0' + date.getMinutes()).slice(-2);

    var str = y + '-' + m + '-' + d + ' ' + h + ':' + mi;

    if (showColor === 'color') {
        if (diff < day) {
            return '<font color="green">' + str + '</font>';
        } else if (diff < 3 * day) {
            return '<font color="blue">' + str + '</font>';
        }
    }
    return str;
}

// long2ip 转换
function long2ip(ip) {
    if (!ip) return '';
    if (typeof ip === 'string' && ip.indexOf('.') > -1) return ip;
    return [
        (ip >>> 24) & 0xFF,
        (ip >>> 16) & 0xFF,
        (ip >>> 8) & 0xFF,
        ip & 0xFF
    ].join('.');
}

// HTML 转义
function htmlspecialchars(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
}

// 编辑器初始化 (简化版)
function editor_getEditor(id) {
    // 返回 textarea 元素的简单包装
    var el = document.getElementById(id);
    return {
        getContent: function() { return el ? el.value : ''; },
        setContent: function(val) { if (el) el.value = val; }
    };
}

function editor_getContent(editor) {
    return editor ? editor.getContent() : '';
}

function editor_setContent(editor, val) {
    if (editor) editor.setContent(val);
}

var ROOT_PATH = '';

// 模块数据 API 映射
var ModuleAPI = {
    // 视频模块
    vod: {
        list: '/admin/api/vod/list',
        detail: '/admin/api/vod/detail',
        save: '/admin/api/vod/save',
        del: '/admin/api/vod/del',
        field: '/admin/api/vod/field'
    },
    // 文章模块
    art: {
        list: '/admin/api/art/list',
        detail: '/admin/api/art/detail',
        save: '/admin/api/art/save',
        del: '/admin/api/art/del',
        field: '/admin/api/art/field'
    },
    // 分类模块
    type: {
        list: '/admin/api/type/list',
        detail: '/admin/api/type/detail',
        save: '/admin/api/type/save',
        del: '/admin/api/type/del'
    },
    // 用户模块
    user: {
        list: '/admin/api/user/list',
        detail: '/admin/api/user/detail',
        save: '/admin/api/user/save',
        del: '/admin/api/user/del'
    },
    // 管理员模块
    admin: {
        list: '/admin/api/admin/list',
        detail: '/admin/api/admin/detail',
        save: '/admin/api/admin/save',
        del: '/admin/api/admin/del'
    },
    // 采集模块
    collect: {
        list: '/admin/api/collect/list',
        detail: '/admin/api/collect/detail',
        save: '/admin/api/collect/save',
        del: '/admin/api/collect/del'
    },
    // 演员模块
    actor: {
        list: '/admin/api/actor/list',
        detail: '/admin/api/actor/detail',
        save: '/admin/api/actor/save',
        del: '/admin/api/actor/del'
    },
    // 角色模块
    role: {
        list: '/admin/api/role/list',
        detail: '/admin/api/role/detail',
        save: '/admin/api/role/save',
        del: '/admin/api/role/del'
    },
    // 专题模块
    topic: {
        list: '/admin/api/topic/list',
        detail: '/admin/api/topic/detail',
        save: '/admin/api/topic/save',
        del: '/admin/api/topic/del'
    },
    // 留言模块
    gbook: {
        list: '/admin/api/gbook/list',
        detail: '/admin/api/gbook/detail',
        save: '/admin/api/gbook/save',
        del: '/admin/api/gbook/del'
    },
    // 评论模块
    comment: {
        list: '/admin/api/comment/list',
        detail: '/admin/api/comment/detail',
        save: '/admin/api/comment/save',
        del: '/admin/api/comment/del'
    },
    // 链接模块
    link: {
        list: '/admin/api/link/list',
        detail: '/admin/api/link/detail',
        save: '/admin/api/link/save',
        del: '/admin/api/link/del'
    },
    // 漫画模块
    manga: {
        list: '/admin/api/manga/list',
        detail: '/admin/api/manga/detail',
        save: '/admin/api/manga/save',
        del: '/admin/api/manga/del'
    },
    // 系统设置
    system: {
        config: '/admin/api/system/config',
        saveConfig: '/admin/api/system/saveConfig'
    }
};
