<s>[{{T .Loc "sitename"}}] Password Reset</s>

<div>
	<p>Hi, {{.User.Name}}</p>
	<p>Thank you for using {{T .Loc "sitename"}}.</p>
	<p>Your login password reset request for &lt;{{.User.Email}}&gt; has been received on {{TIME .Now}}.</p>
	<p>If you want to reset your password, please click the following link within {{.Expires}} minutes.</p>
	<p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
