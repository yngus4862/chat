using System.Text.Json.Serialization;

namespace Chat.Client.Core;

public sealed record Room(
    [property: JsonPropertyName("id")] string Id,
    [property: JsonPropertyName("name")] string Name
);