const thumbWidth = 128, thumbHeight = 128;

// 文件保质期 (变灰时间), 该值应等于后端的 database.turnGrey
const TurnGrey = { n: 15, unit: 'days' };

// 向服务器提交表单，在等待过程中 btn 会失效，避免重复提交。
function ajaxPost(form, url, btn, onload, onloadend) {
  if (btn) {
    btn.prop('disabled', true);
  }
  let xhr = new XMLHttpRequest();

  xhr.responseType = 'json';
  xhr.open('POST', url);

  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  };
  
  xhr.onload = onload;

  xhr.addEventListener('loadend', function() {
    if (btn) {
      btn.prop('disabled', false);
    }
    if (onloadend) onloadend();
  });

  xhr.send(form);
}

// 向服务器提交表单，在等待过程中名为 button_name 的按钮会隐藏，
// 同时同名的 spinner 会出现，暗示用户耐心等待，同时避免重复提交。
function ajaxPostWithSpinner(form, url, button_name, onload, onloadend) {
  if (button_name) {
    $(`#${button_name}-btn`).hide();
    $(`#${button_name}-spinner`).show();
  }

  let xhr = new XMLHttpRequest();
  xhr.responseType = 'json';

  xhr.open('POST', url);
  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  };
  xhr.addEventListener('loadend', function() {
    if (button_name) {
      $(`#${button_name}-btn`).show();
      $(`#${button_name}-spinner`).hide();
    }
    if (onloadend) onloadend();
  });
  xhr.onload = onload;
  xhr.send(form);
}

// 从服务器获取数据。
function ajaxGet(url, btn, onload, onloadend) {
  if (btn) {
    btn.prop('disabled', true);
  }
  let xhr = new XMLHttpRequest();

  xhr.responseType = 'json';
  xhr.open('GET', url);

  xhr.onerror = function () {
    window.alert('An error occurred during the transaction');
  }
  
  xhr.onload = onload;

  xhr.addEventListener('loadend', function() {
    if (btn) {
      btn.prop('disabled', false);
    }
    if (onloadend) onloadend();
  });

  xhr.send();
}

// 插入出错提示
function insertErrorAlert(msg, where) {
  insertAlert('danger', msg, where);
}

// 插入普通提示
function insertInfoAlert(msg, where) {
  insertAlert('info', msg, where);
}

// 插入成功提示
function insertSuccessAlert(msg, where) {
  insertAlert('success', msg, where);
}

// 插入提示
function insertAlert(type, msg, where) {
  console.log(msg);
  let alertElem = $('#alert-'+type+'-tmpl').contents().clone();
  alertElem.find('.AlertMessage').text(msg);
  if (!where) where = '#alert-insert-after-here';
  alertElem.insertAfter(where);
}

// 把文件大小换算为 KB 或 MB
function fileSizeToString(fileSize, fixed) {
  if (fixed == null) {
    fixed = 2
  }
  sizeMB = fileSize / 1024 / 1024;
  if (sizeMB < 1) {
      return `${(sizeMB * 1024).toFixed(fixed)} KB`;
  }
  return `${sizeMB.toFixed(fixed)} MB`;
}

// https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
// In Chrome 60, they added a feature that disables crypto.subtle for non-TLS connections.
async function sha256Hex(file) {
  let buffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  return hashHex;
}

// 缩略图的url
function thumbURL(id) {
  return '/files/' + id + '.small';
}

// 文件的url
function fileURL(id) {
  return '/files/' + id + '.send';
}

// 带时间的url, 用于刷新文件。
function urlWithDate(originURL) {
  let d = new Date();
  return originURL + '?' + d.getTime();
}

// 文件未必是图片，因此尝试生成缩略图，如果出错则说明这不是图片。
async function tryToDrawThumb(file, imgElem) {
  let src = URL.createObjectURL(file);
  await drawThumb(file.type, src, imgElem);
  URL.revokeObjectURL(src);
}

// 生成缩略图显示在 imgElem 里，如果不是 video 就当作是 image.
function drawThumb(filetype, src, imgElem) {
  return new Promise((resolve, reject) => {
    let mediaElem;
    if (filetype.startsWith('video/')) {
      let video = document.createElement('video');
      mediaElem = video;
      video.src = src;
      video.addEventListener('loadeddata', function() {
        // 取第几秒的截图？不想取第一帧（因为片头可能与视频内容关系不大），
        // 也不想取太后面的截图（担心消耗太多资源）。
        video.currentTime = video.duration / 30;
      });
      video.addEventListener('timeupdate', function() {
        let sw = video.videoWidth, sh = video.videoHeight;
        drawToImg(sw, sh, mediaElem, imgElem);
        resolve();
      });
    } else {
      let img = document.createElement('img');
      mediaElem = img;
      img.src = src;
      img.onload = function() {
        let sw = img.width, sh = img.height;
        drawToImg(sw, sh, mediaElem, imgElem);
        resolve();
      }
    }
    mediaElem.onerror = reject;
  });
}

// mediaElem 可能是 img, 也可能是 video.
function drawToImg(sw, sh, mediaElem, imgElem) {
  // 截取原图中间的正方形
  let sx = 0, sy = 0;
  if (sw > sh) {
      sx = (sw - sh) / 2;
      sw = sh;
  } else {
      sy = (sh - sw) / 2;
      sh = sw;
  }
  let canvas = document.createElement('canvas');
  canvas.width = thumbWidth;
  canvas.height = thumbHeight;
  let ctx = canvas.getContext('2d');
  ctx.drawImage(mediaElem, sx, sy, sw, sh, 0, 0, thumbWidth, thumbHeight);
  imgElem.src = canvas.toDataURL('image/jpeg');
}

async function dataUrlToFile(dataUrl, filename) {
  let blob = await fetch(dataUrl).then(r => r.blob());
  return new File([blob], filename, {
    type: 'image/jpeg',
    lastModified: Date.now()
  });
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// ItemID.next() 输出自增 id.
let ItemID = {
  n: 0,
  next: function() {
    this.n++;
    return 'item-' + this.n;
  },
  current: function() {
    return 'item-' + this.n;
  }
};

function getThumbByFile(file) {
  return getThumbByFiletype(typeOfFile(file));
}

function typeOfFile(file) {
  let fileNameParts = file.name.split('.');
  if (fileNameParts.length <= 1) {
    return file.type;
  }
  let ext = fileNameParts.pop();
  let filetype;
  if (["zip", "rar", "7z", "gz", "tar", "bz", "bz2", "xz"].indexOf(ext) >= 0) {
    filetype = "compressed/" + ext;
  } else if (["md", "xml", "html", "xhtml", "htm"].indexOf(ext) >= 0) {
    filetype = "text/" + ext
  } else if (["doc", "docx", "ppt", "pptx", "rtf", "xls", "xlsx"].indexOf(ext) >= 0) {
    filetype = "office/" + ext
  } else if (["epub", "pdf", "mobi", "azw", "azw3", "djvu"].indexOf(ext) >= 0) {
    filetype = "ebook/" + ext
  } else {
    filetype = file.type;
  }
  return filetype
}

function getThumbByFiletype(filetype) {
  let prefix = filetype.split('/').shift();
  let suffix = filetype.split('/').pop();
  switch (suffix) {
    case 'doc':
    case 'docx':
      return '/public/icons/file-earmark-word.jpg';
    case 'xls':
    case 'xlsx':
      return '/public/icons/file-earmark-excel.jpg';
    case 'ppt':
    case 'pptx':
      return '/public/icons/file-earmark-ppt.jpg';
    default:
      switch (prefix) {
        case 'office':
        case 'ebook':
          return '/public/icons/file-earmark-richtext.jpg';
        case 'compressed':
          return '/public/icons/file-earmark-zip.jpg';
        case 'text':
          return '/public/icons/file-earmark-text.jpg';
        case 'audio':
          return '/public/icons/file-earmark-music.jpg';
        default:
          return '/public/icons/file-earmark-binary.jpg';
      }    
  }
}



// go-send-demo 专用
async function drawThumbResize(file, imgElem) {
  let src = URL.createObjectURL(file);
  await drawThumb(file.type, src, imgElem);
  let [canvas, changed] = await resizeLimit(src, null);

  // 如果图片不需要缩小，就返回 null.
  if (!changed) {
    return [null, false];
  }

  let blob = await canvasToJPEG(canvas);
  let img_resized = new File([blob], file.name + '.jpeg', {
    type: 'image/jpeg',
    lastModified: Date.now()
  });
  URL.revokeObjectURL(src);
  return [img_resized, true];
}

// Convert `canvas.toBlob` to promise style.
// 本来想转换为 webp, 但有些浏览器不支持, 而且转换速度很慢。
function canvasToJPEG(canvas) {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      canvas.toBlob = null;
      reject(Error('timeout'));
    }, 3000); // timeout: 3 seconds

    canvas.toBlob(blob => {
      clearTimeout(timeout);
      resolve(blob);
    }, 'image/jpeg');
  });
}

// ResizeLimit resizes the src if it's long side bigger than limit.
// Use default limit if limit is set to zero or null.
function resizeLimit(src, limit) {
  return new Promise((resolve, reject) => {
    let img = document.createElement('img');
    img.src = src;
    img.onload = function() {
      let [dw, dh] = limitWidthHeight(img.width, img.height, limit);

      // 如果图片小于限制值，其大小就保持不变。
      if (dw == img.width && dh == img.height) {
        resolve([null, false]);
      }

      let canvas = document.createElement('canvas');
      canvas.width = thumbWidth;
      canvas.height = thumbHeight;
      let ctx = canvas.getContext('2d');
      ctx.drawImage(img, 0, 0, img.width, img.height, 0, 0, dw, dh);
      resolve([canvas, true]);
    };
    img.onerror = reject;
  });
}

function limitWidthHeight(w, h, limit) {
  if (!limit) {
    limit = 900 // 默认边长上限 900px
  }
  // 先限制宽度
  if (w > limit) {
    h *= limit / w
    w = limit
  }
  // 缩小后的高度仍有可能超过限制，因此要再判断一次
  if (h > limit) {
    w *= limit / h
    h = limit
  }
  return [w, h];
}
