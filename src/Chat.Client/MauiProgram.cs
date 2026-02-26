using Chat.Client.Core.Abstractions;
using Chat.Client.Infrastructure.Auth;
using Chat.Client.Infrastructure.Notifications;
using Chat.Client.Infrastructure.Options;
using Chat.Client.Infrastructure.Realtime;
using Chat.Client.Infrastructure.Rest;
using Chat.Client.Infrastructure.Storage;
using Chat.Client.ViewModels;
using CommunityToolkit.Mvvm.Messaging;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;

namespace Chat.Client;

public static class MauiProgram
{
    public static MauiApp CreateMauiApp()
    {
        var builder = MauiApp.CreateBuilder();

        builder
            .UseMauiApp<App>()
            .ConfigureFonts(fonts =>
            {
                fonts.AddFont("OpenSans-Regular.ttf", "OpenSansRegular");
                fonts.AddFont("OpenSans-Semibold.ttf", "OpenSansSemibold");
            });

        // Configuration: appsettings.json + env
        builder.Configuration
            .AddJsonFile("appsettings.json", optional: false, reloadOnChange: true)
            .AddEnvironmentVariables(prefix: "CHAT__");

#if DEBUG
        builder.Logging.AddDebug();
#endif
        builder.Services.AddSingleton<AppShell>();

        // Options
        builder.Services.Configure<ApiOptions>(builder.Configuration.GetSection("Api"));
        builder.Services.Configure<RealtimeOptions>(builder.Configuration.GetSection("Realtime"));
        builder.Services.Configure<AuthOptions>(builder.Configuration.GetSection("Auth"));

        // App state + messenger
        builder.Services.AddSingleton<IAppState, SimpleAppState>();
        builder.Services.AddSingleton<IMessenger>(WeakReferenceMessenger.Default);

        // Storage / Auth
        builder.Services.AddSingleton<ITokenStore, SecureTokenStore>();
        builder.Services.AddSingleton<IAuthService, OidcAuthService>();

        // REST
        builder.Services.AddHttpClient<IChatApi, ChatApiClient>();

        // Realtime
        builder.Services.AddSingleton<IRealtimeClient, RealtimeWsClient>();

        // Notifications (platform-specific)
#if WINDOWS
        builder.Services.AddSingleton<INotificationService, Chat.Client.Platforms.Windows.WindowsToastNotificationService>();
        builder.Services.AddSingleton<Chat.Client.Platforms.Windows.WindowsAppNotificationsBootstrapper>();
#else
        builder.Services.AddSingleton<INotificationService, NullNotificationService>();
#endif
        builder.Services.AddSingleton<NotificationThrottler>();

        // ViewModels
        builder.Services.AddTransient<RoomsViewModel>();
        builder.Services.AddTransient<ChatViewModel>();

        // Pages
        builder.Services.AddTransient<Pages.RoomsPage>();
        builder.Services.AddTransient<Pages.ChatPage>();

        return builder.Build();
    }
}

public sealed class SimpleAppState : IAppState
{
    public string? ActiveRoomId { get; set; }
    public bool IsChatPageVisible { get; set; }
}