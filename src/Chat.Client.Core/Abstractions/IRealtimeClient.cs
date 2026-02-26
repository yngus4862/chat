namespace Chat.Client.Core.Abstractions;

public interface IRealtimeClient : IAsyncDisposable
{
    event EventHandler<ChatMessage>? MessageReceived;

    Task ConnectRoomAsync(string roomId, CancellationToken ct);
    Task DisconnectAsync(CancellationToken ct);
    bool IsConnected { get; }
}