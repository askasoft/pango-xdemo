[{{HTML (T .Loc "site")}}] 登录验证码的通知

<div>
	<p>感谢您使用 {{T .Loc "site"}}。</p>
	<p>您在 {{TIME .Now}} 正在登录网站。</p>
	<p>{{.Expires}}分钟以内，请在登录页面输入以下的验证码。</p>
	<p>验证码: {{.Passcode}}</p>
	<br>
	<p>此致,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
