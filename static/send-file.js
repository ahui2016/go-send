// 如果有些函数在这里找不到，那就是在 util.js 里。

// eruda.init();
let files = [];

// 上传文件
$('#send-btn').click(event => {
    event.preventDefault();
    if (files.length == 0) {
        insertErrorAlert('未选择文件');
        return;
    }
    hideForm();
    files.forEach(file => setCardAlert('info', file.itemID, 'Waiting...'));
    uploadOneByOne(files.length);
});

// 文件逐个上传
function uploadOneByOne(i) {
    if (i <= 0) {
        insertInfoAlert('上传结束，如果如下所示。');
        refreshTotalSize();
        return;
    }
    i--;
    checkHashUpload(i);
}

let fileInput = $('#file-input');
fileInput.change(event => {

    $('#hidden-area').show();

    for (const file of event.target.files) {

        const sameFile = a => a.name === file.name;

        let i = files.findIndex(sameFile);
        if (i > -1) {
            insertInfoAlert(`已忽略重复文件: ${file.name}`);
            continue;
        }
        files.push(file)
        updateFilesCount(files.length);

        // 插入卡片。图片或视频有缩略图
        let item = $('#file-msg-tmpl').contents().clone()
        item.insertAfter('#file-msg-tmpl');
        if (file.type.startsWith('image/') || file.type.startsWith('video/')) {
            tryToDrawThumb(file, item.find('.card-img')[0]);
        } else {
            let thumbUrl = getThumbByFile(file);
            item.find('.card-img').attr('src', thumbUrl);
        }

        // 填充卡片内容（文件名等）
        let itemID = ItemID.next();
        file.itemID = itemID;
        item.attr('id', itemID);
        item.find('.FileSize').text(fileSizeToString(file.size));
        item.find('.card-text').text(file.name);
        item.find('.Icon').tooltip();

        // 删除按钮
        let deleteIcon = item.find('.DeleteIcon');
        deleteIcon.click(() => {
            let i = files.findIndex(sameFile);
            files.splice(i, 1);
            updateFilesCount(files.length);
            deleteIcon.tooltip('dispose');
            item.hide('slow', function(){ item.remove() });
        });
    }
});

// 更新已选择的文件数量
function updateFilesCount(count) {
    let plural = ' file';
    if (count > 1) {
        plural = ' files';
    }
    $('#files-count').text(count + plural);
}

// 检查文件的校验和，如无冲突则上传文件。
async function checkHashUpload(i) {
    let file = files[i];
    setCardAlert('info', file.itemID, 'Uploading...');
    let fileSha256 = await sha256Hex(file);

    let form = new FormData();
    form.append('hashHex', fileSha256);

    ajaxPost(form, '/api/checksum', $('#upload-btn'), function() {
        if (this.status == 200) {
            uploadFile(fileSha256, i);
        } else {
            let errMsg = !this.response ? this.status : this.response.message;
            setCardDanger(file.itemID, 'Error: ' + errMsg);
            uploadOneByOne(i);
        }
    });
}

// 上传文件
async function uploadFile(fileSha256, i) {
    let file = files[i];
    let form = new FormData();
    form.append('file', file);
    form.append('checksum', fileSha256);
    form.append('filename', file.name);
    form.append('filesize', file.size);

    if (file.type.startsWith('video/')) {
        let dataUrl = $('#' + file.itemID).find('img')[0].src;
        let thumbFile = await dataUrlToFile(dataUrl, file.itemID + '.jpg');
        form.append('thumbnail', thumbFile);
    }

    ajaxPostWithSpinner(form, '/api/upload-file', 'upload', function() {
            if (this.status == 200) {
                setCardSuccess(file.itemID, 'OK. File Uploaded.')
            } else if (this.status == 413) {
                let errMsg = 'File Too Large';
                setCardDanger(file.itemID, 'Error: ' + errMsg);
            } else {
                let errMsg = !this.response ? this.status : this.response.message;
                setCardDanger(file.itemID, 'Error: ' + errMsg);
            }
        },
        function() {
            uploadOneByOne(i);
        });
}

function setCardDanger(itemID, msg) {
    setCardAlert('danger', itemID, msg);
}

function setCardSuccess(itemID, msg) {
    setCardAlert('success', itemID, msg);
}

function setCardAlert(type, itemID, msg) {
    let cardItem = $('#'+itemID);
    cardItem.removeClass('border-info');
    cardItem.addClass('border-' + type);
    cardItem.find('.IconButtons').hide();
    // cardItem.find('.card-subtitle').hide();
    cardItem.find('.ResultMsg')
        .removeClass('text-info')
        .addClass('text-' + type)
        .text(msg)
        .show();
}

function hideForm() {
    $('#custom-file-input').hide()
    $('#hidden-area').hide();
}

// Safari(iOS) has a bug.
// https://gist.github.com/hanayashiki/8dac237671343e7f0b15de617b0051bd
(function () {
    File.prototype.arrayBuffer = File.prototype.arrayBuffer || myArrayBuffer;
    Blob.prototype.arrayBuffer = Blob.prototype.arrayBuffer || myArrayBuffer;

    function myArrayBuffer() {
        // this: File or Blob
        return new Promise((resolve) => {
            let fr = new FileReader();
            fr.onload = () => {
                resolve(fr.result);
            };
            fr.readAsArrayBuffer(this);
        })
    }
})();

refreshTotalSize();
function refreshTotalSize() {
    ajaxGet("/api/total-size", null, function() {
        if (this.status == 200) {
            let 已用 = fileSizeToString(this.response.totalSize, 0);
            let 剩余可用 = fileSizeToString(
                this.response.capacity - this.response.totalSize, 0);
            $('#total-size').text(`已用: ${已用}, 剩余可用: ${剩余可用}`);
        } else {
            let errMsg = !this.response ? this.status : this.response.message;
            console.log(errMsg);
        }
    });
}

$('.NavbarBtn').tooltip();

fileInput.focus();
