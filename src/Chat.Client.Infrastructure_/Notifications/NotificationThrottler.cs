namespace Chat.Client.Infrastructure.Notifications;

public sealed class NotificationThrottler
{
    private readonly TimeSpan _window;
    private readonly int _burstCount;

    private readonly Dictionary<string, (DateTimeOffset start, int count)> _state = new();

    public NotificationThrottler(int windowSeconds = 8, int burstCount = 3)
    {
        _window = TimeSpan.FromSeconds(windowSeconds);
        _burstCount = burstCount;
    }

    // returns: (shouldShow, forceSummary, count)
    public (bool show, bool summary, int count) Hit(string roomId)
    {
        var now = DateTimeOffset.UtcNow;

        if (!_state.TryGetValue(roomId, out var s) || now - s.start > _window)
            s = (now, 0);

        s.count++;
        _state[roomId] = s;

        if (s.count == 1) return (true, false, 1);
        if (s.count >= _burstCount) return (true, true, s.count);

        // 2번째부터는 “요약”으로 갈지/그냥 무시할지 정책 선택 가능
        return (false, false, s.count);
    }
}