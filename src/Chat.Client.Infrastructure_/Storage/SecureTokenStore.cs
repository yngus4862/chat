using Microsoft.Maui.Storage;
using Chat.Client.Core.Abstractions;

namespace Chat.Client.Infrastructure.Storage;

public sealed class SecureTokenStore : ITokenStore
{
    private const string AccessKey = "chat_access_token";
    private const string RefreshKey = "chat_refresh_token";
    private const string ExpiresKey = "chat_expires_at";

    public Task<string?> GetAccessTokenAsync() => SecureStorage.GetAsync(AccessKey);

    public async Task SetTokensAsync(string accessToken, string? refreshToken, DateTimeOffset expiresAt)
    {
        await SecureStorage.SetAsync(AccessKey, accessToken);
        if (!string.IsNullOrWhiteSpace(refreshToken))
            await SecureStorage.SetAsync(RefreshKey, refreshToken);

        await SecureStorage.SetAsync(ExpiresKey, expiresAt.ToUnixTimeSeconds().ToString());
    }

    public Task ClearAsync()
    {
        SecureStorage.Remove(AccessKey);
        SecureStorage.Remove(RefreshKey);
        SecureStorage.Remove(ExpiresKey);
        return Task.CompletedTask;
    }
}