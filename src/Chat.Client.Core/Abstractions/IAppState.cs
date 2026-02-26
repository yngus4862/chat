namespace Chat.Client.Core.Abstractions;

public interface IAppState
{
    string? ActiveRoomId { get; set; }
    bool IsChatPageVisible { get; set; }
}