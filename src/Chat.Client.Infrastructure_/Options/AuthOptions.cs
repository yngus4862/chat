namespace Chat.Client.Infrastructure.Options;

public sealed class AuthOptions
{
    public bool Enabled { get; set; } = false;
    public string Issuer { get; set; } = "";
    public string ClientId { get; set; } = "";
    public string RedirectUri { get; set; } = "mychat://callback";
    public string Scope { get; set; } = "openid profile email";
}