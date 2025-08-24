<s>[{{T .Loc "sitename"}}] 密码重置</s>

<div>
	<p>你好, {{.User.Name}}</p>
	<p>感谢您使用 {{T .Loc "sitename"}}。</p>
	<p>您在 {{TIME .Now}} 申请了&lt;{{.User.Email}}&gt;的登录密码重置。</p>
	<p>如果您想重置密码，请在{{.Expires}}分钟内点击下面的链接。</p>
	<p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
	<br>
	<p>此致,</p>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
