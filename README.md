# biscuit
 hand-rolled cookies library in go

This library is for setting and deleting specifically login cookies (although other types of cookies
will be rolled out in the future). The library includes a simple sessions manager to keep track of things like logged in users, but note that more advanced use cases should roll their own sessions manager or use a more robust manager than the one provided in this package