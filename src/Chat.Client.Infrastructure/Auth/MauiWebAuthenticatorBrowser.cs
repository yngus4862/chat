using Duende.IdentityModel.OidcClient.Browser;

#if WINDOWS
using WinUIEx;
#else
using Microsoft.Maui.Authentication;
#endif

namespace Chat.Client.Infrastructure.Auth;

public sealed class MauiWebAuthenticatorBrowser : Duende.IdentityModel.OidcClient.Browser.IBrowser
{
    public async Task<BrowserResult> InvokeAsync(BrowserOptions options, CancellationToken cancellationToken = default)
    {
        try
        {
#if WINDOWS
            var result = await WinUIEx.WebAuthenticator.AuthenticateAsync(
                new Uri(options.StartUrl),
                new Uri(options.EndUrl));
#else
            var result = await WebAuthenticator.AuthenticateAsync(
                new Uri(options.StartUrl),
                new Uri(options.EndUrl));
#endif

            var query = string.Join("&", result.Properties.Select(kv =>
                $"{Uri.EscapeDataString(kv.Key)}={Uri.EscapeDataString(kv.Value)}"));

            var responseUrl = options.EndUrl + (options.EndUrl.Contains("?") ? "&" : "?") + query;

            return new BrowserResult { ResultType = BrowserResultType.Success, Response = responseUrl };
        }
        catch (TaskCanceledException)
        {
            return new BrowserResult { ResultType = BrowserResultType.UserCancel };
        }
        catch (Exception ex)
        {
            return new BrowserResult { ResultType = BrowserResultType.UnknownError, Error = ex.ToString() };
        }
    }
}