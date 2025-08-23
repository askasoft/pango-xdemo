[{{HTML (T .Loc "site")}}] Password Reset

<div>
	<p>Hi, {{.User.Name}}</p>
	<p>Thank you for using {{T .Loc "site"}}.</p>
	<p>Your login password reset request for &lt;{{.User.Email}}&gt; has been received on {{TIME .Now}}.</p>
	<p>If you want to reset your password, please click the following link within {{.Expires}} minutes.</p>
	<p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
