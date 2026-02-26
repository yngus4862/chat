#if WINDOWS
using Chat.Client.Core.Abstractions;
using Microsoft.Windows.AppNotifications;
using Microsoft.Windows.AppNotifications.Builder;

namespace Chat.Client.Platforms.Windows;

public sealed class WindowsToastNotificationService : INotificationService
{
    private readonly WindowsAppNotificationsBootstrapper _boot;

    public WindowsToastNotificationService(WindowsAppNotificationsBootstrapper boot)
    {
        _boot = boot;
        _boot.EnsureRegistered();
    }

    public Task ShowMessageToastAsync(string roomName, string sender, string previewText, string deepLink)
    {
        _boot.EnsureRegistered();

        var b = new AppNotificationBuilder()
            .AddText($"{sender}  ·  {roomName}")
            .AddText(previewText)
            .AddArgument("link", deepLink);

        var n = b.BuildNotification();
        AppNotificationManager.Default.Show(n);

        return Task.CompletedTask;
    }
}
#endif