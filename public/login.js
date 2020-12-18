// 如果有些函数在这里找不到，那就是在 util.js 里。

const loginBtn = $('#login-btn');
loginBtn.click(event => {
    event.preventDefault();
    let password = $('#password').val();
    if (password == '') {
        insertInfoAlert('请输入密码');
        return;
    }

    let form = new FormData();
    form.append('password', password);

    ajaxPostWithSpinner(form, '/login', 'login', function() {
        if (this.status == 200) {
            loginBtn.prop('disabled', true);
            window.location = '/home';
        } else {
            let errMsg = !this.response ? this.status : this.response.message;
            insertErrorAlert(errMsg);
        }
    });
});
