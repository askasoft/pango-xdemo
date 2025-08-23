[{{HTML (T .Loc "site")}}] Security Notice

<div>
	<p>Hi, {{.User.Name}}</p>
	<p>Thank you for using {{T .Loc "site"}}.</p>
	<p>Your login password for &lt;{{.User.Email}}&gt; was reset on {{TIME .Now}}.</p>
	<p>If you are doing it yourself, you don't need to do anything.</p>
	<p>If you are not aware of any reset operations, please change your login password immediately.</p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
