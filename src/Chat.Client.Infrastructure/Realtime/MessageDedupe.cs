using System.Security.Cryptography;
using System.Text;

namespace Chat.Client.Infrastructure.Realtime;

public sealed class MessageDedupe
{
    private readonly int _capacity;
    private readonly LinkedList<string> _order = new();
    private readonly HashSet<string> _seen = new();

    public MessageDedupe(int capacity = 2000) => _capacity = capacity;

    public bool TryMarkSeen(string key)
    {
        if (_seen.Contains(key)) return false;

        _seen.Add(key);
        _order.AddLast(key);

        while (_order.Count > _capacity)
        {
            var first = _order.First!.Value;
            _order.RemoveFirst();
            _seen.Remove(first);
        }
        return true;
    }

    public static string StableKey(string roomId, string? id, string? sender, string content, DateTimeOffset? createdAt)
    {
        if (!string.IsNullOrWhiteSpace(id)) return $"{roomId}:{id}";

        var raw = $"{roomId}|{sender}|{createdAt?.ToUnixTimeMilliseconds()}|{content}";
        var bytes = SHA256.HashData(Encoding.UTF8.GetBytes(raw));
        return Convert.ToHexString(bytes);
    }
}