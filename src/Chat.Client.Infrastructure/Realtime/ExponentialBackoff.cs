namespace Chat.Client.Infrastructure.Realtime;

public sealed class ExponentialBackoff
{
    private readonly TimeSpan _min;
    private readonly TimeSpan _max;
    private readonly Random _rng = new();

    public ExponentialBackoff(TimeSpan? min = null, TimeSpan? max = null)
    {
        _min = min ?? TimeSpan.FromMilliseconds(400);
        _max = max ?? TimeSpan.FromSeconds(10);
    }

    public TimeSpan NextDelay(int attempt)
    {
        var pow = Math.Pow(2, Math.Clamp(attempt, 0, 8));
        var ms = _min.TotalMilliseconds * pow;

        // jitter 0.7~1.3
        ms *= 0.7 + _rng.NextDouble() * 0.6;
        ms = Math.Min(ms, _max.TotalMilliseconds);
        return TimeSpan.FromMilliseconds(ms);
    }
}