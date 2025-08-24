<s>[{{T .Loc "sitename"}}] 登录验证码的通知</s>

<div>
	<p>感谢您使用 {{T .Loc "sitename"}}。</p>
	<p>您在 {{TIME .Now}} 正在登录网站。</p>
	<p>{{.Expires}}分钟以内，请在登录页面输入以下的验证码。</p>
	<p>验证码: {{.Passcode}}</p>
	<br>
	<p>此致,</p>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
