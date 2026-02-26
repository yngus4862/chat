using System.Text.Json.Serialization;

namespace Chat.Client.Core;

public sealed record ChatMessage
{
    [JsonPropertyName("id")] public string? Id { get; init; }
    [JsonPropertyName("roomId")] public string? RoomId { get; init; }
    [JsonPropertyName("sender")] public string? Sender { get; init; }
    [JsonPropertyName("content")] public string Content { get; init; } = "";
    [JsonPropertyName("createdAt")] public DateTimeOffset? CreatedAt { get; init; }

    // 서버가 필드를 더 보내도 깨지지 않게
    [JsonExtensionData] public Dictionary<string, object>? Extra { get; init; }
}