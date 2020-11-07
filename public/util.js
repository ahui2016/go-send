const thumbWidth = 128, thumbHeight = 128;

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
function fileSizeToString(fileSize) {
  sizeMB = fileSize / 1024 / 1024;
  if (sizeMB < 1) {
      return `${(sizeMB * 1024).toFixed(2)} KB`;
  }
  return `${sizeMB.toFixed(2)} MB`;
}

// 把标签文本框内的字符串转化为数组。
function getNewTags() {
  let trimmed = $('#tags-input').val().replace(/#|,|，/g, ' ').trim();
  if (trimmed.length == 0) {
    return [];
  }
  return trimmed.split(/ +/);
}

// 把标签数组转化为字符串。
function addPrefix(arr, prefix) {
  if (arr == null) {
    return '';
  }
  return arr.map(x => prefix + x).join(' ');
}

// 检查服务器中有无文件冲突（内容完全一样的文件）。
function checkHashHex(hashHex) {
  let form = new FormData();
  form.append('hashHex', hashHex);
  ajaxPost(form, '/api/checksum', null, function() {
    if (this.status == 200) {
      console.log('OK');
    } else {
      console.log(`Error: ${this.status} ${JSON.stringify(this.response)}`);
    }
  });
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

// 日期：月和日
function monthAndDay(simpledatetime) {
  return simpledatetime.split(' ').slice(0, 2).join(' ');
}

// 日期：年月日和时间
function simpleDateTime(date) {
  return date.toString().split(' ').slice(1, 5).join(' ');
}

// 日期：年月日
function simpleDate(date) {
  let year = '' + date.getFullYear(),
      month = '' + (date.getMonth() + 1),
      day = '' + date.getDate();
  if (month.length < 2) month = '0' + month;
  if (day.length < 2) day = '0' + day;
  return [year, month, day].join('-');
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

// Convert `FileReader.readAsDataURL` to promise style.
function readFilePromise(file) {
  return new Promise((resolve, reject) => {
    let reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => {
      resolve(reader.result);
    };
    reader.onerror = reject;
  });
}

// 文件未必是图片，因此尝试生成缩略图，如果出错则说明这不是图片。
async function tryToDrawThumb(file, canvasElem, imgElem) {
  try {
    let src = await readFilePromise(file);
    await drawThumb(src, canvasElem, imgElem);
    // canvasToImg(canvasElem[0], imgElem[0]);
  } catch (e) {
    console.log(e);
  }
}

// https://developer.mozilla.org/en-US/docs/Web/API/HTMLCanvasElement/toBlob
function canvasToImg(canvas, img) {
  canvas.toBlob(blob => {
    let url = URL.createObjectURL(blob);
    img.onload = function() {
      URL.revokeObjectURL(url);
    }
    img.src = url;
  });
}

// 生成缩略图并描绘到一个 canvas 里。
function drawThumb(src, canvasElem, imgElem) {
  return new Promise((resolve, reject) => {
    let img = new Image();
    img.src = src;
    img.onload = function() {

      // 截取原图中间的正方形
      let sw = img.width, sh = img.height;
      let sx = 0, sy = 0;
      if (sw > sh) {
          sx = (sw - sh) / 2;
          sw = sh;
      } else {
          sy = (sh - sw) / 2;
          sh = sw;
      }

      canvasElem
        .attr('width', thumbWidth)
        .attr('height', thumbHeight);
      let ctx = canvasElem[0].getContext('2d'); // canvasElem[0] is the raw html-element.
      ctx.drawImage(img, sx, sy, sw, sh, 0, 0, thumbWidth, thumbHeight);

      imgElem[0].src = canvasElem[0].toDataURL();
      resolve();
    };
    img.onerror = reject;
  });
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// 将 arr 里的 box (ID == box_id) 移动到顶部，并保持其他元素的顺序。
function moveBoxToTop(box_id, tail) {
  let i = tail.findIndex(box => box.ID == box_id);
  if (i < 0) return null;
  if (i == 0) return tail;
  if (i == 1) {
    [tail[0], tail[1]] = [tail[1], tail[0]];
    return tail;
  }

  // 特殊情况如上。以下是普通情况。
  let head = tail.splice(i, 1);
  head.push(...tail);
  return head;
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
