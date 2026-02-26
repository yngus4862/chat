namespace Chat.Client.Core.Abstractions;

public interface IAuthService
{
    Task<bool> IsEnabledAsync();
    Task<bool> IsSignedInAsync();
    Task SignInAsync(CancellationToken ct);
    Task SignOutAsync(CancellationToken ct);
}