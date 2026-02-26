using Chat.Client.Core;
using Chat.Client.Core.Abstractions;
using Chat.Client.Infrastructure.Notifications;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System.Collections.ObjectModel;

namespace Chat.Client.ViewModels;

public partial class ChatViewModel : ObservableObject
{
    private readonly IChatApi _api;
    private readonly IRealtimeClient _rt;
    private readonly INotificationService _noti;
    private readonly NotificationThrottler _throttle;
    private readonly IAppState _state;

    public ObservableCollection<ChatMessage> Messages { get; } = new();

    [ObservableProperty] private string title = "";
    [ObservableProperty] private string draft = "";

    public string RoomId { get; private set; } = "";
    public string RoomName { get; private set; } = "Room";

    public ChatViewModel(
        IChatApi api,
        IRealtimeClient rt,
        INotificationService noti,
        NotificationThrottler throttle,
        IAppState state)
    {
        _api = api;
        _rt = rt;
        _noti = noti;
        _throttle = throttle;
        _state = state;

        _rt.MessageReceived += OnRealtimeMessage;
    }

    public void SetRoom(string roomId, string? roomName)
    {
        RoomId = Uri.UnescapeDataString(roomId ?? "");
        RoomName = string.IsNullOrWhiteSpace(roomName) ? "Room" : Uri.UnescapeDataString(roomName);
        Title = $"#{RoomName} ({RoomId})";
    }

    public async Task OpenAsync()
    {
        if (string.IsNullOrWhiteSpace(RoomId)) return;

        Messages.Clear();
        var list = await _api.GetMessagesAsync(RoomId, limit: 50, CancellationToken.None);
        foreach (var m in list) Messages.Add(m);

        await _rt.ConnectRoomAsync(RoomId, CancellationToken.None);
    }

    public Task CloseAsync() => _rt.DisconnectAsync(CancellationToken.None);

    [RelayCommand]
    public async Task SendAsync()
    {
        var text = Draft?.Trim();
        if (string.IsNullOrWhiteSpace(text)) return;

        Draft = "";
        var sent = await _api.SendMessageAsync(RoomId, text, CancellationToken.None);

        // optimistic UI
        Messages.Add(sent with { RoomId = RoomId, Sender = sent.Sender ?? "me" });
    }

    private void OnRealtimeMessage(object? sender, ChatMessage msg)
    {
        if (msg.RoomId != RoomId) return;

        MainThread.BeginInvokeOnMainThread(() => Messages.Add(msg));

        var isForegroundAndViewing = _state.IsChatPageVisible && _state.ActiveRoomId == RoomId;
        if (isForegroundAndViewing) return;

        var (show, summary, count) = _throttle.Hit(RoomId);
        if (!show) return;

        var preview = summary ? $"새 메시지 {count}개" : msg.Content;
        var deepLink = $"mychat://room/{Uri.EscapeDataString(RoomId)}";

        _ = _noti.ShowMessageToastAsync(RoomName, msg.Sender ?? "unknown", preview, deepLink);
    }
}