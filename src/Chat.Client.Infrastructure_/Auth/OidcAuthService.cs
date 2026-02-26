using Duende.IdentityModel.OidcClient;
using Duende.IdentityModel.Client;
using Chat.Client.Core.Abstractions;
using Chat.Client.Infrastructure.Options;
using Microsoft.Extensions.Options;

namespace Chat.Client.Infrastructure.Auth;

public sealed class OidcAuthService : IAuthService
{
    private readonly AuthOptions _opt;
    private readonly ITokenStore _tokenStore;

    public OidcAuthService(IOptions<AuthOptions> opt, ITokenStore tokenStore)
    {
        _opt = opt.Value;
        _tokenStore = tokenStore;
    }

    public Task<bool> IsEnabledAsync() => Task.FromResult(_opt.Enabled);

    public async Task<bool> IsSignedInAsync()
    {
        var token = await _tokenStore.GetAccessTokenAsync();
        return !string.IsNullOrWhiteSpace(token);
    }

    public async Task SignInAsync(CancellationToken ct)
    {
        if (!_opt.Enabled) return;

        var oidc = new OidcClient(new OidcClientOptions
        {
            Authority = _opt.Issuer.TrimEnd('/'),
            ClientId = _opt.ClientId,
            Scope = _opt.Scope,
            RedirectUri = _opt.RedirectUri,
            Browser = new MauiWebAuthenticatorBrowser(),
            Policy = new Policy
            {
                Discovery = new DiscoveryPolicy
                {
                    // 사내망/개발환경에서 HTTPS가 아닐 수도 있어서 (PROD는 true 권장)
                    RequireHttps = _opt.Issuer.StartsWith("https://", StringComparison.OrdinalIgnoreCase)
                }
            }
        });

        var login = await oidc.LoginAsync(new LoginRequest(), ct);
        if (login.IsError)
            throw new InvalidOperationException($"OIDC login failed: {login.Error}");

        var expiresAt = DateTimeOffset.UtcNow.AddSeconds(login.TokenResponse.ExpiresIn);
        await _tokenStore.SetTokensAsync(login.AccessToken, login.RefreshToken, expiresAt);
    }

    public async Task SignOutAsync(CancellationToken ct)
    {
        await _tokenStore.ClearAsync();
        await Task.CompletedTask;
    }
}