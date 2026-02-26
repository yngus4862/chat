namespace Chat.Client;

public partial class AppShell : Shell
{
	public AppShell()
	{
		InitializeComponent();

        // 라우트 등록 (ChatPage QueryProperty로 roomId/roomName 받음)
        Routing.RegisterRoute("chat", typeof(Pages.ChatPage));
    }
}
