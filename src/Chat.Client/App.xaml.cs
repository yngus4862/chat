#if WINDOWS
using Chat.Client.Platforms.Windows;
#endif

using Microsoft.Maui.Controls;

namespace Chat.Client;

public partial class App : Application
{
    private readonly AppShell _shell;

    public App(
        AppShell shell
#if WINDOWS
        , WindowsAppNotificationsBootstrapper boot
#endif
        )
    {
        InitializeComponent();

        _shell = shell;

#if WINDOWS
        // Windows App SDK 알림 등록(토스트 사용 전에 1회)
        boot.EnsureRegistered();
#endif
    }

    protected override Window CreateWindow(IActivationState? activationState)
    {
        return new Window(_shell);
    }
}