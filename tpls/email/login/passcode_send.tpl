[{{HTML (T .Loc "site")}}] Notice about two-factor email authentication

<div>
	<p>Thank you for using {{T .Loc "site"}}.</p>
	<p>A login request has been received at {{TIME .Now}}.</p>
	<p>Please continue the two-step authentication within {{.Expires}} minutes.</p>
	<p>Passcode: {{.Passcode}}</p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
