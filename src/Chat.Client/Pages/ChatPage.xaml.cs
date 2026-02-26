using Chat.Client.Core.Abstractions;
using Chat.Client.ViewModels;

namespace Chat.Client.Pages;

[QueryProperty(nameof(RoomId), "roomId")]
[QueryProperty(nameof(RoomName), "roomName")]
public partial class ChatPage : ContentPage
{
    private readonly IAppState _state;
    private readonly ChatViewModel _vm;

    public ChatPage(ChatViewModel vm, IAppState state)
    {
        InitializeComponent();
        _vm = vm;
        _state = state;
        BindingContext = _vm;
    }

    public string RoomId
    {
        set => _vm.SetRoom(value, RoomName);
    }

    public string RoomName { get; set; } = "Room";

    protected override async void OnAppearing()
    {
        base.OnAppearing();
        _state.IsChatPageVisible = true;
        _state.ActiveRoomId = _vm.RoomId;

        await _vm.OpenAsync();
    }

    protected override async void OnDisappearing()
    {
        base.OnDisappearing();
        _state.IsChatPageVisible = false;
        _state.ActiveRoomId = null;

        await _vm.CloseAsync();
    }
}