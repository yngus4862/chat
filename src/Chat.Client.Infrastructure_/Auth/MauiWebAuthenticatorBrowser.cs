using Duende.IdentityModel.OidcClient.Browser;
using Microsoft.Maui.Authentication;

namespace Chat.Client.Infrastructure.Auth;

// OidcClient가 “시스템 브라우저”를 열기 위한 브릿지
public sealed class MauiWebAuthenticatorBrowser : IBrowser
{
    public async Task<BrowserResult> InvokeAsync(BrowserOptions options, CancellationToken cancellationToken = default)
    {
        try
        {
            //var result = await WebAuthenticator.AuthenticateAsync(
            //    new Uri(options.StartUrl),
            //    new Uri(options.EndUrl));
            var authUri = new Uri(options.StartUrl);
            var redirectUri = new Uri(options.EndUrl);
#if WINDOWS
            var result = await WinUIEx.WebAuthenticator.AuthenticateAsync(authUri, redirectUri);
#else
            var result = await WebAuthenticator.AuthenticateAsync(authUri, redirectUri);
#endif

            // WebAuthenticator는 쿼리 파라미터 컬렉션을 제공
            var query = string.Join("&", result.Properties.Select(kv =>
                $"{Uri.EscapeDataString(kv.Key)}={Uri.EscapeDataString(kv.Value)}"));

            var responseUrl = options.EndUrl + (options.EndUrl.Contains("?") ? "&" : "?") + query;

            return new BrowserResult
            {
                ResultType = BrowserResultType.Success,
                Response = responseUrl
            };
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