using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using Chat.Client.Core;
using Chat.Client.Core.Abstractions;
using Chat.Client.Infrastructure.Options;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Chat.Client.Infrastructure.Rest;

public sealed class ChatApiClient : IChatApi
{
    private readonly HttpClient _http;
    private readonly ITokenStore _tokenStore;
    private readonly ApiOptions _opt;
    private readonly ILogger<ChatApiClient> _log;

    private static readonly JsonSerializerOptions JsonOpt = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public ChatApiClient(
        HttpClient http,
        IOptions<ApiOptions> opt,
        ITokenStore tokenStore,
        ILogger<ChatApiClient> log)
    {
        _http = http;
        _opt = opt.Value;
        _tokenStore = tokenStore;
        _log = log;

        _http.BaseAddress = new Uri(_opt.BaseUrl.TrimEnd('/') + "/");
        _http.Timeout = TimeSpan.FromSeconds(30);
    }

    private async Task AttachAuthAsync()
    {
        var token = await _tokenStore.GetAccessTokenAsync();
        _http.DefaultRequestHeaders.Authorization =
            string.IsNullOrWhiteSpace(token) ? null : new AuthenticationHeaderValue("Bearer", token);
    }

    public async Task<IReadOnlyList<Room>> GetRoomsAsync(CancellationToken ct)
    {
        await AttachAuthAsync();
        using var res = await _http.GetAsync("v1/rooms", ct);
        res.EnsureSuccessStatusCode();

        var json = await res.Content.ReadAsStringAsync(ct);
        return JsonSerializer.Deserialize<List<Room>>(json, JsonOpt) ?? [];
    }

    public async Task<Room> CreateRoomAsync(string name, CancellationToken ct)
    {
        await AttachAuthAsync();

        var body = JsonSerializer.Serialize(new { name });
        using var res = await _http.PostAsync("v1/rooms",
            new StringContent(body, Encoding.UTF8, "application/json"), ct);

        res.EnsureSuccessStatusCode();
        var json = await res.Content.ReadAsStringAsync(ct);

        // 서버가 {id,name}를 준다고 가정. 필드명이 다르면 여기만 맞추면 됨.
        return JsonSerializer.Deserialize<Room>(json, JsonOpt)
               ?? throw new InvalidOperationException("CreateRoom 응답 파싱 실패");
    }

    public async Task<IReadOnlyList<ChatMessage>> GetMessagesAsync(string roomId, int limit, CancellationToken ct)
    {
        await AttachAuthAsync();
        using var res = await _http.GetAsync($"v1/rooms/{Uri.EscapeDataString(roomId)}/messages?limit={limit}", ct);
        res.EnsureSuccessStatusCode();

        var json = await res.Content.ReadAsStringAsync(ct);
        return JsonSerializer.Deserialize<List<ChatMessage>>(json, JsonOpt) ?? [];
    }

    public async Task<ChatMessage> SendMessageAsync(string roomId, string content, CancellationToken ct)
    {
        await AttachAuthAsync();

        var body = JsonSerializer.Serialize(new { content });
        using var res = await _http.PostAsync($"v1/rooms/{Uri.EscapeDataString(roomId)}/messages",
            new StringContent(body, Encoding.UTF8, "application/json"), ct);

        res.EnsureSuccessStatusCode();
        var json = await res.Content.ReadAsStringAsync(ct);

        // 서버가 메시지 객체 반환한다고 가정. 아니면 {ok:true} 같은 경우 여기만 조정.
        return JsonSerializer.Deserialize<ChatMessage>(json, JsonOpt)
               ?? new ChatMessage { RoomId = roomId, Content = content, CreatedAt = DateTimeOffset.UtcNow };
    }
}