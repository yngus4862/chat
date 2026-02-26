namespace Chat.Client.Infrastructure.Options;

public sealed class RealtimeOptions
{
    public string WsBaseUrl { get; set; } = "ws://localhost:8081";
    public string Path { get; set; } = "/ws";
}