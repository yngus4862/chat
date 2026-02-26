#if WINDOWS
using Microsoft.Windows.AppNotifications;

namespace Chat.Client.Platforms.Windows;

public sealed class WindowsAppNotificationsBootstrapper
{
    private bool _registered;

    public void EnsureRegistered()
    {
        if (_registered) return;

        // (문서 권장) NotificationInvoked 핸들러 등록 -> Register 순서 :contentReference[oaicite:14]{index=14}
        AppNotificationManager.Default.NotificationInvoked += (_, args) =>
        {
            // args.Argument / args.UserInput
            // 여기서 deep link 처리(예: Shell navigate)까지 확장 가능
        };

        AppNotificationManager.Default.Register();
        _registered = true;
    }

    public void Unregister()
    {
        if (!_registered) return;
        AppNotificationManager.Default.Unregister();
        _registered = false;
    }
}
#endif