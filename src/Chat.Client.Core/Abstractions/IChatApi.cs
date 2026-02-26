namespace Chat.Client.Core.Abstractions;

public interface IChatApi
{
    Task<IReadOnlyList<Room>> GetRoomsAsync(CancellationToken ct);
    Task<Room> CreateRoomAsync(string name, CancellationToken ct);

    Task<IReadOnlyList<ChatMessage>> GetMessagesAsync(string roomId, int limit, CancellationToken ct);
    Task<ChatMessage> SendMessageAsync(string roomId, string content, CancellationToken ct);
}