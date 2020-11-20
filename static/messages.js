// 如果有些函数在这里找不到，那就是在 public/util.js 里。

const msgInput = $('#msg-input');
const commandHelp = $('#command-help');
const commands = $('#commands');
const executeBtn = $('#execute-btn')
const page = $('.navbar-brand').text();

initData();

function initData() {
    let url;
    if (page == 'Messages') url = '/api/all';
    if (page == 'Clips') url = '/api/all-clips';
    ajaxGet(url, null, function() {
            if (this.status == 200) {

                // 条目数太少时不显示高级功能，Clips 页面也不显示高级功能
                if (this.response.length >= 5) {
                    $('#commands-form').show();
                }

                this.response.forEach(message => {

                    // 两种类型的不同操作
                    let item;
                    if (message.Type == 'TextMsg') {
                        item = insertTextMsg(message);
                    } else if (message.Type == 'FileMsg') {
                        item = insertFileMsg(message);
                    } else {
                        insertErrorAlert('Unknown message type: ' + message.Type);
                    }

                    // 两种类型的相同操作
                    doAfterInsert(item, message);
                });
            } else {
                let errMsg = !this.response ? this.status : this.response.message;
                insertErrorAlert(errMsg);
            }
        },
        function() {
            $('#loading-spinner').hide();
        });
}

// 两种类型的相同操作
function doAfterInsert(item, message) {

    // 通用按钮
    let iconButtons = $('#icon_buttons').contents().clone();
    iconButtons.insertAfter(item.find('.通用按钮插入位置'));

    let itemID = 'item-'+message.ID;
    item.attr('id', itemID);

    let simple_id = simpleID(message.ID);
    item.find('.MsgID').text(simple_id);
    item.find('.Icon').tooltip();

    // 如果当前时间超过保质期（当前时间在变灰时间之后），该卡片就会变灰。变灰表示已过期，即将被自动删除。
    let expired = dayjs(message.UpdatedAt).add(TurnGrey.n, TurnGrey.unit);
    if (dayjs().isAfter(expired)) {
        item.addClass('bg-light');
        item.find('.InfoIcon').show();
        if (message.Type == 'TextMsg') item.find('.CopyIcon').hide();
        if (message.Type == 'FileMsg') item.find('.DownloadIcon').hide();
    }

    // 顶置按钮
    let up_button = item.find('.UpIcon');
    if (page == 'Clips') up_button.hide();
    up_button.click(() => {
        let form = new FormData();
        form.append('id', message.ID);
        up_button.hide();
        ajaxPost(form, '/api/update-datetime', null, function() {
                if (this.status == 200) {
                    up_button.tooltip('hide');
                    $('html').animate({scrollTop:0}, 50);
                    item.removeClass('bg-light');
                    item.find('.InfoIcon').hide();
                    if (message.Type == 'TextMsg') item.find('.CopyIcon').show();
                    if (message.Type == 'FileMsg') item.find('.DownloadIcon').show();
                    // 插入时，要么插在 #file-msg-tmpl 后面，要么插在 #text-msg-tmpl 前面。
                    item.insertAfter('#file-msg-tmpl');
                } else {
                    let errMsg = !this.response ? this.status : this.response.message;
                    insertErrorAlert(errMsg, '#'+itemID);
                }
            },
            function() {
                up_button.show();
            });
    });

    // 删除按钮
    let delete_button = item.find('.DeleteIcon');
    delete_button.click(() => {
        delete_button.tooltip('hide');
        $('#delete-dialog').modal('show');
        $('#id-in-modal').text(simple_id);

        if (message.Type == 'TextMsg') {
            $('#confirm-question').text('Delete this message?');
            $('#filesize-in-modal').text('');
        }
        if (message.Type == 'FileMsg') {
            $('#confirm-question').text('Delete this file?');
            $('#filesize-in-modal').text('('+fileSizeToString(message.FileSize)+')');
        }

        // 确认删除
        $('#yes-button').off().click(() => {
            let url;
            if (page == 'Messages') url = '/api/delete';
            if (page == 'Clips') url = '/api/delete-clip';
            let form = new FormData();
            form.append('id', message.ID);
            ajaxPost(form, url, $('#yes-button'), function() {
                    if (this.status == 200) {
                        let infoMsg = 'id: ' + simple_id + ' is deleted';
                        item.hide('slow', function() {
                            insertInfoAlert(infoMsg, '#'+itemID);
                            item.remove();
                        });
                    } else {
                        window.alert(`Error: ${this.status} ${this.statusText} ${this.response}`);
                    }
                },
                function() {
                    $('#delete-dialog').modal('hide');
                });
        });
    });
}

function insertTextMsg(message) {
    const item = $('#text-msg-tmpl').contents().clone();
    // 插入时，要么插在 #file-msg-tmpl 后面，要么插在 #text-msg-tmpl 前面。
    item.insertAfter('#file-msg-tmpl');
    item.find('.card-text').text(message.TextMsg);


    // 复制按钮
    const copyIcon = item.find('.CopyIcon');
    const copybtnID = 'copybtn-'+message.ID;
    copyIcon.attr('id', copybtnID);
    const clipboard = new ClipboardJS('#'+copybtnID, {
        text: function(){
            return message.TextMsg;
        }
    });
    clipboard.on('success', () => {
        copyIcon
            .tooltip('dispose')
            .attr('title', 'copied!')
            .tooltip('show');
        window.setTimeout(function() {
            copyIcon.tooltip('dispose').attr('title', 'copy').tooltip();
        }, 2000);
    });
    clipboard.on('error', e => {
        console.error('Action:', e.action);
        console.error('Trigger:', e.trigger);
    });

    return item;
}

function insertFileMsg(message) {
    let item = $('#file-msg-tmpl').contents().clone();
    // 插入时，要么插在 #file-msg-tmpl 后面，要么插在 #text-msg-tmpl 前面。
    item.insertAfter('#file-msg-tmpl');

    let file = {name: message.FileName, type: message.FileType};
    if (file.type.startsWith('image/') || file.type.startsWith('video/')) {
        item.find('.card-img')
            .on('error', event => event.currentTarget.src = '/public/icons/file-earmark-play.jpg' )
            .attr('src', thumbURL(message.ID));
        item.find('.LinkToBigImg').attr('href', fileURL(message.ID));
        item.find('.col-md-2').addClass('col-3');
    } else {
        let thumbUrl = getThumbByFile(file);
        item.find('.card-img').attr('src', thumbUrl);
        item.find('.col-md-2').addClass('col-2');
    }

    item.find('.card-text').text(message.FileName);
    item.find('.FileSize').text(fileSizeToString(message.FileSize));
    item.find('.DownloadButton')
        .attr('href', fileURL(message.ID))
        .attr('download', message.FileName);

    return item;
}

// simpleID 取正常 id 的最后三个字符作为简化 id, 方便人眼辨认。
function simpleID(id) {
    let len = id.length;
    return id.slice(len-3, len);
}

// 发送文字备忘
$('#send-btn').click(sendMsg);
function sendMsg(event) {
    event.preventDefault();

    let msg = msgInput.val().trim();
    if (msg == '') {
        msgInput.focus();
        return;
    }

    let form = new FormData();
    form.append('text-msg', msg);
    ajaxPostWithSpinner(form, '/api/add-text-msg', 'send', function() {
        if (this.status == 200) {
            let message = this.response;
            let item = insertTextMsg(message);
            doAfterInsert(item, message);
            msgInput.val('');
        } else {
            let errMsg = !this.response ? this.status : this.response.message;
            insertErrorAlert(errMsg);
        }
        msgInput.focus();
    });
}

// 提供高级命令的说明
commands.change(event => {
    switch (event.currentTarget.value) {
        case 'zip-all-files':
            commandHelp.text('打包全部文件，不包括文字备忘。打包后，压缩包会显示在列表顶部。下载后请尽快删除以节省空间。');
            break;
        case 'delete-all-files':
            commandHelp.text('删除全部文件，保留删除文字备忘。删除后本页面会自动刷新，被删除的文件不可恢复。');
            break;
        case 'delete-10-files':
            commandHelp.text('删除列表底部 10 个文件，保留文字备忘。如果在列表底部有不想删除的文件，可点击其 “上升” 按钮使其上升至列表顶部，但如果一共只有 10 个文件或更少，则全部文件都会被删除。');
            break;
        case 'delete-10-items':
            commandHelp.text('删除列表底部 10 项，包括文件和文字备忘，被删除的项目不可恢复。');
            break;
        case 'delete-grey-items':
            commandHelp.text('删除已变灰的项目(过期项目)，若有不想删除的项目可点击 “上升” 按钮使其上升至列表顶部，被删除的项目不可恢复。');
            break;
        default:
            commandHelp.text('');
    }
});

// 执行高级命令
executeBtn.click(event => {
    event.preventDefault();

    let command = commands.val();
    if (command == "none") {
        insertInfoAlert('请从下拉菜单选择后再执行。', $('#all-messages'));
        commands.focus();
        return;
    }

    let form = new FormData();
    form.append('command', command);
    ajaxPostWithSpinner(form, '/api/execute-command', 'execute', function() {
        if (this.status == 200) {
            // 如果 this.response 是一个 Message, 则插入列表顶部，否则刷新页面。
            if (this.response && this.response.ID) {
                $('html').animate({scrollTop:0},'50');
                let message = this.response;
                let item = insertFileMsg(message);
                doAfterInsert(item, message);
                return;
            }
            executeBtn.prop('disabled', true);
            window.location.reload();
        } else {
            let errMsg = !this.response ? this.status : this.response.message;
            insertErrorAlert(errMsg, $('#all-messages'));
            commands.focus();
        }
    });
});

$('.NavbarBtn').tooltip();

msgInput.focus();
