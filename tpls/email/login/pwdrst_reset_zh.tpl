[{{HTML (T .Loc "site")}}] 安全通知

<div>
	<p>你好, {{.User.Name}}</p>
	<p>感谢您使用 {{T .Loc "site"}}。</p>
	<p>&lt;{{.User.Email}}&gt;的登录密码已于 {{TIME .Now}} 被重置。</p>
	<p>如果是您自己重置的，则无需执行任何操作。</p>
	<p>如果您对重置操作不确定，请立即更改您的登录密码。</p>
	<br>
	<p>此致,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
