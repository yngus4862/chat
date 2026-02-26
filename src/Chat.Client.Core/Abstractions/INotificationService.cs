namespace Chat.Client.Core.Abstractions;

public interface INotificationService
{
    Task ShowMessageToastAsync(string roomName, string sender, string previewText, string deepLink);
}