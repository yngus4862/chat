using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Chat.Client.Core;
using Chat.Client.Core.Abstractions;
using Chat.Client.Infrastructure.Options;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Chat.Client.Infrastructure.Realtime;

public sealed class RealtimeWsClient : IRealtimeClient
{
    private readonly RealtimeOptions _opt;
    private readonly ILogger<RealtimeWsClient> _log;

    private readonly ExponentialBackoff _backoff = new();
    private readonly MessageDedupe _dedupe = new();

    private ClientWebSocket? _ws;
    private CancellationTokenSource? _loopCts;
    private string? _roomId;

    public event EventHandler<ChatMessage>? MessageReceived;

    public bool IsConnected => _ws?.State == WebSocketState.Open;

    public RealtimeWsClient(IOptions<RealtimeOptions> opt, ILogger<RealtimeWsClient> log)
    {
        _opt = opt.Value;
        _log = log;
    }

    public async Task ConnectRoomAsync(string roomId, CancellationToken ct)
    {
        _roomId = roomId;

        // 기존 루프 종료
        if (_loopCts is not null)
        {
            _loopCts.Cancel();
            _loopCts.Dispose();
        }

        _loopCts = CancellationTokenSource.CreateLinkedTokenSource(ct);
        _ = Task.Run(() => RunLoopAsync(_loopCts.Token), _loopCts.Token);
        await Task.CompletedTask;
    }

    public async Task DisconnectAsync(CancellationToken ct)
    {
        _loopCts?.Cancel();
        if (_ws is { State: WebSocketState.Open or WebSocketState.CloseReceived })
        {
            try { await _ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "client disconnect", ct); }
            catch { /* ignore */ }
        }
        _ws?.Dispose();
        _ws = null;
    }

    private async Task RunLoopAsync(CancellationToken ct)
    {
        if (string.IsNullOrWhiteSpace(_roomId)) return;

        var attempt = 0;

        while (!ct.IsCancellationRequested)
        {
            try
            {
                await EnsureConnectedAsync(_roomId!, ct);
                attempt = 0;

                await ReceiveLoopAsync(ct);
            }
            catch (OperationCanceledException)
            {
                return;
            }
            catch (Exception ex)
            {
                _log.LogWarning(ex, "WS loop error. will retry.");
                await SafeDisposeWsAsync();
                var delay = _backoff.NextDelay(attempt++);
                try { await Task.Delay(delay, ct); } catch { return; }
            }
        }
    }

    private async Task EnsureConnectedAsync(string roomId, CancellationToken ct)
    {
        if (_ws?.State == WebSocketState.Open) return;

        await SafeDisposeWsAsync();

        var url = BuildWsUrl(roomId);
        _log.LogInformation("WS connecting: {Url}", url);

        _ws = new ClientWebSocket();
        _ws.Options.KeepAliveInterval = TimeSpan.FromSeconds(20);

        await _ws.ConnectAsync(url, ct);
        _log.LogInformation("WS connected.");
    }

    private Uri BuildWsUrl(string roomId)
    {
        var baseUrl = _opt.WsBaseUrl.TrimEnd('/');
        var path = _opt.Path.StartsWith("/") ? _opt.Path : "/" + _opt.Path;
        var full = $"{baseUrl}{path}?roomId={Uri.EscapeDataString(roomId)}";
        return new Uri(full);
    }

    private async Task ReceiveLoopAsync(CancellationToken ct)
    {
        if (_ws is null) return;

        var buffer = new byte[64 * 1024];

        while (!ct.IsCancellationRequested && _ws.State == WebSocketState.Open)
        {
            var sb = new StringBuilder();
            WebSocketReceiveResult? result;

            do
            {
                result = await _ws.ReceiveAsync(buffer, ct);

                if (result.MessageType == WebSocketMessageType.Close)
                    throw new WebSocketException("WS closed by server.");

                sb.Append(Encoding.UTF8.GetString(buffer, 0, result.Count));
            }
            while (!result.EndOfMessage);

            var text = sb.ToString().Trim();
            if (string.IsNullOrWhiteSpace(text)) continue;

            if (TryParseMessage(text, _roomId!, out var msg))
            {
                var key = MessageDedupe.StableKey(_roomId!, msg.Id, msg.Sender, msg.Content, msg.CreatedAt);
                if (_dedupe.TryMarkSeen(key))
                    MessageReceived?.Invoke(this, msg);
            }
        }
    }

    private static bool TryParseMessage(string json, string roomId, out ChatMessage msg)
    {
        msg = new ChatMessage { RoomId = roomId, Content = json };

        try
        {
            using var doc = JsonDocument.Parse(json);
            var root = doc.RootElement;

            // 최소: {"content":"hello"} (README의 wscat 예시) :contentReference[oaicite:9]{index=9}
            if (root.ValueKind == JsonValueKind.Object && root.TryGetProperty("content", out var c))
            {
                msg = new ChatMessage
                {
                    RoomId = root.TryGetProperty("roomId", out var r) ? r.GetString() : roomId,
                    Id = root.TryGetProperty("id", out var id) ? id.GetString() : null,
                    Sender = root.TryGetProperty("sender", out var s) ? s.GetString() : "unknown",
                    Content = c.GetString() ?? "",
                    CreatedAt = root.TryGetProperty("createdAt", out var t) && t.ValueKind == JsonValueKind.String
                        ? DateTimeOffset.TryParse(t.GetString(), out var dto) ? dto : null
                        : null
                };
                return true;
            }

            return false;
        }
        catch
        {
            // JSON이 아니면 그냥 텍스트 메시지로 처리
            msg = new ChatMessage { RoomId = roomId, Sender = "unknown", Content = json, CreatedAt = DateTimeOffset.UtcNow };
            return true;
        }
    }

    private Task SafeDisposeWsAsync()
    {
        try { _ws?.Dispose(); } catch { /* ignore */ }
        _ws = null;
        return Task.CompletedTask;
    }

    public async ValueTask DisposeAsync()
    {
        try { _loopCts?.Cancel(); } catch { }
        _loopCts?.Dispose();
        _loopCts = null;

        await SafeDisposeWsAsync();
    }
}