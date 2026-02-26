namespace Chat.Client.Core.Abstractions;

public interface ITokenStore
{
    Task<string?> GetAccessTokenAsync();
    Task SetTokensAsync(string accessToken, string? refreshToken, DateTimeOffset expiresAt);
    Task ClearAsync();
}