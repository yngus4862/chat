using Chat.Client.Core.Abstractions;

namespace Chat.Client.Infrastructure.Notifications;

public sealed class NullNotificationService : INotificationService
{
    public Task ShowMessageToastAsync(string roomName, string sender, string previewText, string deepLink)
        => Task.CompletedTask;
}